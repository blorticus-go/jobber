package jobber

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Runner struct {
	client          *Client
	config          *Configuration
	resourceTracker *CreatedResourceTracker
}

func NewRunner(config *Configuration, client *Client) *Runner {
	return &Runner{
		client:          client,
		config:          config,
		resourceTracker: NewCreatedResourceTracker(),
	}
}

func (runner *Runner) createDefaultNamespace(pipelineVariables *PipelineVariables) (*corev1.Namespace, error) {
	action, err := PipelineActionFromStringDescriptor("resources/default-namespace.yaml", runner.config.Test.Pipeline.ActionDefinitionsRootDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to create action for resources/default-namespace.yaml: %s", err)
	}

	actionEventChannel := make(chan *ActionEvent)
	go action.Run(pipelineVariables, &PipelineExecutionEnvironment{EnvironmentalVariables: runner.config.Test.Pipeline.ExecutionEnvironment}, runner.client, actionEventChannel)

	var createdResource *GenericK8sResource

EventLoop:
	for {
		event := <-actionEventChannel
		switch event.Type {
		case AnErrorOccurred:
			return nil, fmt.Errorf("error on attempt to create default namespace from resources/default-namespace.yaml: %s", err)
		case ResourceCreated:
			createdResource = event.AffectedResource
		case ActionCompletedSuccessfully:
			break EventLoop
		}
	}

	nsObject := new(corev1.Namespace)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(createdResource.ApiObject().Object, nsObject); err != nil {
		return nil, fmt.Errorf("failed to convert unstructured namespace object to typed: %s", err)
	}

	namespaceName := nsObject.Name

	runner.resourceTracker.AddCreatedResource(&DeletableK8sResource{
		information: &K8sResourceInformation{
			Kind:          "Namespace",
			Name:          namespaceName,
			NamespaceName: "",
		},
		deletionMethod: func(object any) error {
			return runner.client.DeleteNamespace(namespaceName)
		},
	})

	return nsObject, nil
}

func (runner *Runner) RunTest(eventChannel chan<- *Event) {
	eventHandler := &eventHandler{eventChannel}
	assetsDirectoryManager := NewContextualAssetsDirectoryManager()

	outcome := assetsDirectoryManager.CreateTestAssetsRootDirectory()
	if eventHandler.explainAssetCreationOutcome(outcome, nil, nil); outcome.DirectoryCreationFailureError != nil {
		return
	}

	templateExpansionVariables := NewEmptyPipelineVariables(runner.client).WithGlobalValues(runner.config.Test.GlobalValues)

	testCasePipeline, err := NewPipelineFromStringDescriptors(runner.config.Test.Pipeline.ActionsInOrder, runner.config.Test.Pipeline.ActionDefinitionsRootDirectory)
	if err != nil {
		eventHandler.sayThatPipelineDefinitionIsInvalid(err)
		return
	}

	for _, testUnit := range runner.config.Test.Units {
		eventHandler.sayThatUnitStarted(testUnit)

		outcome := assetsDirectoryManager.CreateTestUnitDirectory(testUnit)
		if eventHandler.explainAssetCreationOutcome(outcome, testUnit, nil); outcome.DirectoryCreationFailureError != nil {
			return
		}

		templateExpansionVariables := templateExpansionVariables.RescopedToUnitNamed(testUnit.Name).WithUnitValues(testUnit.Values)

		for _, testCase := range runner.config.Test.Cases {
			eventHandler.sayThatCaseStarted(testUnit, testCase)

			outcome := assetsDirectoryManager.CreateTestCaseDirectories(testUnit, testCase)
			if eventHandler.explainAssetCreationOutcome(outcome, testUnit, testCase); outcome.DirectoryCreationFailureError != nil {
				return
			}

			nsObject, err := runner.createDefaultNamespace(templateExpansionVariables)
			if eventHandler.explainAttemptToCreateDefaultNamespace(nsObject, EventContextFor(testUnit, testCase), err); err != nil {
				return
			}

			templateExpansionVariables := templateExpansionVariables.
				RescopedToCaseNamed(testCase.Name).
				WithCaseValues(testCase.Values).
				AndTestCaseRetrievedAssetsDirectoryAt(assetsDirectoryManager.TestCaseAssetsDirectoryPathsFor(testUnit, testCase).RetrievedAssets).
				AndUsingDefaultNamespaceNamed(nsObject.Name)

			for action := testCasePipeline.Restart(); action != nil; action = testCasePipeline.NextAction() {
				actionEventChannel := make(chan *ActionEvent)

				go action.Run(templateExpansionVariables, &PipelineExecutionEnvironment{EnvironmentalVariables: runner.config.Test.Pipeline.ExecutionEnvironment}, runner.client, actionEventChannel)

				if err := runner.handleActionEvents(action, actionEventChannel, eventHandler, assetsDirectoryManager, testUnit, testCase); err != nil {
					return
				}
			}

			for _, attemptDetails := range runner.resourceTracker.AttemptToDeleteAllAsYetUndeletedResources() {
				if attemptDetails.Error != nil {
					eventHandler.sayThatResourceDeletionFailed(attemptDetails.Resource.information, attemptDetails.Error, testUnit, testCase)
					return
				}

				eventHandler.sayThatResourceDeletionSucceeded(attemptDetails.Resource.information, testUnit, testCase)
			}

			eventHandler.sayThatCaseCompletedSuccessfully(testUnit, testCase)
		}

		eventHandler.sayThatUnitCompletedSuccessfully(testUnit)
	}

	if err := assetsDirectoryManager.GenerateArchiveFileAt(runner.config.Test.AssetArchive.FilePath); err != nil {
		eventHandler.sayThatArchiveCreationFailed(runner.config.Test.AssetArchive.FilePath, assetsDirectoryManager.TestRootAssetDirectoryPath(), err)
		return
	}

	eventHandler.sayThatArchiveCreationSucceeded(runner.config.Test.AssetArchive.FilePath)

	if err := assetsDirectoryManager.RemoveAssetsDirectory(); err != nil {
		eventHandler.sayThatAssetDirectoryDeletionFailed(assetsDirectoryManager.TestRootAssetDirectoryPath(), err)
		return
	}

	eventHandler.sayThatAssetDirectoryDeletionWasSuccessful(assetsDirectoryManager.TestRootAssetDirectoryPath())

	eventHandler.sayThatTestingCompletedSuccessfully()
}

func (runner *Runner) handleActionEvents(action *PipelineAction, actionEventChannel <-chan *ActionEvent, eventHandler *eventHandler, assetsDirectoryManager *ContextualAssetsDirectoryManager, testUnit *TestUnit, testCase *TestCase) error {
	for {
		event := <-actionEventChannel
		switch event.Type {
		case TemplateExpanded:
			writeExpandedTemplateForAction(action, event.ExpandedTemplateBuffer, assetsDirectoryManager.TestCaseAssetsDirectoryPathsFor(testUnit, testCase).ExpandedTemplates)
		case ResourceCreated:
			eventHandler.sayThatResourceCreationSucceeded(event.AffectedResource.Information(), func() string { return "" }, testUnit, testCase)
			runner.resourceTracker.AddCreatedResource(&DeletableK8sResource{
				information: event.AffectedResource.Information(),
				deletionMethod: func(object any) error {
					return event.AffectedResource.Delete()
				},
			})
			switch event.AffectedResource.GvkString() {
			case "v1/Pod":
			case "batch/v1/Job":
			}
		case JobCompleted:
		case PodMovedToRunningState:
		case ExecutionSuccessful:
			attemptToWriteExecutableOutputToFile(assetsDirectoryManager.TestCaseAssetsDirectoryPathsFor(testUnit, testCase).Executables, action.Descriptor, event.StdoutBuffer, event.StderrBuffer)
			eventHandler.sayThatExecutionSucceeded(action.Descriptor, testUnit, testCase)
		case ValuesTransformCompleted:
		case AnErrorOccurred:
			switch action.Type {
			case TemplatedResource:
				if event.AffectedResource != nil {
					eventHandler.sayThatResourceCreationFailed(event.AffectedResource.Information(), func() string { return "" }, event.Error, testUnit, testCase)
				} else {
					eventHandler.sayThatResourceTemplateExpansionFailed(action.ActionFullyQualifiedPath, func() string { return "" }, event.Error, testUnit, testCase)
				}
			case Executable:
				attemptToWriteExecutableOutputToFile(assetsDirectoryManager.TestCaseAssetsDirectoryPathsFor(testUnit, testCase).Executables, action.Descriptor, event.StdoutBuffer, event.StderrBuffer)
				eventHandler.sayThatExecutionFailed(action.Descriptor, event.Error, testUnit, testCase)
			}
			return event.Error
		case ActionCompletedSuccessfully:
			return nil
		}
	}
}

func attemptToWriteExecutableOutputToFile(executableAssetsBasePath string, actionDescriptor string, stdoutBuffer *bytes.Buffer, stderrBuffer *bytes.Buffer) {
	outputFilesBasePath := deriveActionOutputFilesBasePath(executableAssetsBasePath, actionDescriptor)
	writeReaderToFile(fmt.Sprintf("%s.stdout", outputFilesBasePath), 0640, stdoutBuffer)
	writeReaderToFile(fmt.Sprintf("%s.stderr", outputFilesBasePath), 0640, stderrBuffer)
}

func writeExpandedTemplateForAction(action *PipelineAction, expandedTemplateBuffer *bytes.Buffer, assetsDirectoryPath string) {
	outputFilesBasePath := deriveActionOutputFilesBasePath(assetsDirectoryPath, action.ActionFullyQualifiedPath)
	if expandedTemplateBuffer != nil {
		writeReaderToFile(outputFilesBasePath, 0640, expandedTemplateBuffer)
	}
}

func deriveActionOutputFilesBasePath(depositDirectoryPath string, actionFullyQualifiedName string) string {
	actionFullyQalifiedNamePathElements := strings.Split(actionFullyQualifiedName, "/")
	actionBasename := actionFullyQalifiedNamePathElements[len(actionFullyQalifiedNamePathElements)-1]

	basePath := fmt.Sprintf("%s/%s", depositDirectoryPath, actionBasename)
	candidatePath := basePath

	for discriminatorInt := 0; fileExists(candidatePath); discriminatorInt++ {
		candidatePath = fmt.Sprintf("%s-%d", basePath, discriminatorInt)
	}

	return candidatePath
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)

	switch {
	case err == nil:
		return true
	case errors.Is(err, os.ErrNotExist):
		return false
	default:
		panic(fmt.Sprintf("os.Stat failed: %s", err))
	}
}
