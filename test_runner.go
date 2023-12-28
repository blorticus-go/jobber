package jobber

// import (
// 	"errors"
// 	"fmt"
// 	"os"
// 	"strings"

// 	"github.com/blorticus-go/jobber/wrapped"
// 	corev1 "k8s.io/api/core/v1"
// 	"k8s.io/apimachinery/pkg/runtime"
// )

// type TemplateExpansionNamespace struct {
// 	GeneratedName string
// }

// type VariablesConfigContext struct {
// 	TestUnitName                         string
// 	TestCaseName                         string
// 	TestCaseRetrievedAssetsDirectoryPath string
// }

// type TemplateExpansionConfigVariables struct {
// 	DefaultNamespace *TemplateExpansionNamespace
// 	Context          *VariablesConfigContext
// }

// type Runner struct {
// 	client          *Client
// 	config          *Configuration
// 	resourceTracker *wrapped.ResourceTracker
// }

// func NewRunner(config *Configuration, client *Client) *Runner {
// 	return &Runner{
// 		client:          client,
// 		config:          config,
// 		resourceTracker: wrapped.NewResourceTracker(),
// 	}
// }

// func (runner *Runner) createDefaultNamespace(pipelineVariables *PipelineVariables) (*corev1.Namespace, error) {
// 	action, err := PipelineActionFromStringDescriptor("resources/default-namespace.yaml", runner.config.Test.Definition.PipelineRootDirectory)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create action for resources/default-namespace.yaml: %s", err)
// 	}

// 	outcome := (action.Run(pipelineVariables, runner.client))[0]

// 	if outcome.Error != nil {
// 		return nil, fmt.Errorf("error on attempt to create default namespace from resources/default-namespace.yaml: %s", err)
// 	}

// 	nsObject := new(corev1.Namespace)
// 	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(outcome.CreatedResource.ApiObject().Object, nsObject); err != nil {
// 		return nil, fmt.Errorf("failed to convert unstructured namespace object to typed: %s", err)
// 	}

// 	namespaceName := nsObject.Name

// 	runner.resourceTracker.AddCreatedResource(&DeletableK8sResource{
// 		information: &K8sResourceInformation{
// 			Kind:          "Namespace",
// 			Name:          namespaceName,
// 			NamespaceName: "",
// 		},
// 		deletionMethod: func(object any) error {
// 			return runner.client.DeleteNamespace(namespaceName)
// 		},
// 	})

// 	return nsObject, nil
// }

// func (runner *Runner) RunTest(eventChannel chan<- *Event) {
// 	eventHandler := &eventHandler{eventChannel}
// 	assetsDirectoryManager := NewContextualAssetsDirectoryManager()

// 	outcome := assetsDirectoryManager.CreateTestAssetsRootDirectory()
// 	if eventHandler.explainAssetCreationOutcome(outcome, nil, nil); outcome.DirectoryCreationFailureError != nil {
// 		return
// 	}

// 	templateExpansionVariables := NewPipelineVariablesWithSeedValues(runner.config.Test.Definition.DefaultValues, runner.client)

// 	testCasePipeline, err := NewPipelineFromStringDescriptors(runner.config.Test.Definition.Pipeline, runner.config.Test.Definition.PipelineRootDirectory)
// 	if err != nil {
// 		eventHandler.sayThatPipelineDefinitionIsInvalid(err)
// 		return
// 	}

// 	for _, testUnit := range runner.config.Test.Units {
// 		eventHandler.sayThatUnitStarted(testUnit)

// 		outcome := assetsDirectoryManager.CreateTestUnitDirectory(testUnit)
// 		if eventHandler.explainAssetCreationOutcome(outcome, testUnit, nil); outcome.DirectoryCreationFailureError != nil {
// 			return
// 		}

// 		templateExpansionVariables := templateExpansionVariables.MergeValuesToCopy(testUnit.Values)

// 		for _, testCase := range runner.config.Test.Cases {
// 			eventHandler.sayThatCaseStarted(testUnit, testCase)

// 			outcome := assetsDirectoryManager.CreateTestCaseDirectories(testUnit, testCase)
// 			if eventHandler.explainAssetCreationOutcome(outcome, testUnit, testCase); outcome.DirectoryCreationFailureError != nil {
// 				return
// 			}

// 			templateExpansionVariables.Config.Context = &VariablesConfigContext{
// 				TestUnitName:                         testUnit.Name,
// 				TestCaseName:                         testCase.Name,
// 				TestCaseRetrievedAssetsDirectoryPath: assetsDirectoryManager.TestCaseAssetsDirectoryPathsFor(testUnit, testCase).RetrievedAssets,
// 			}

// 			templateExpansionVariables := templateExpansionVariables.MergeValuesToCopy(testCase.Values)

// 			nsObject, err := runner.createDefaultNamespace(templateExpansionVariables)
// 			if eventHandler.explainAttemptToCreateDefaultNamespace(nsObject, EventContextFor(testUnit, testCase), err); err != nil {
// 				return
// 			}

// 			templateExpansionVariables.AddDefaultNamespaceToConfig(nsObject.Name)

// 			for action := testCasePipeline.Restart(); action != nil; action = testCasePipeline.NextAction() {
// 				for _, outcome := range action.Run(templateExpansionVariables, runner.client) {
// 					writeActionOutcomeInformationToAssetsDirectory(action, outcome, assetsDirectoryManager.TestCaseAssetsDirectoryPathsFor(testUnit, testCase))

// 					if eventHandler.explainActionOutcome(action, outcome, testUnit, testCase); outcome.Error != nil {
// 						return
// 					}
// 					if outcome.CreatedResource != nil {
// 						runner.resourceTracker.AddCreatedResource(&DeletableK8sResource{
// 							information: (outcome.CreatedResource.Information()),
// 							deletionMethod: func(object any) error {
// 								return outcome.CreatedResource.Delete()
// 							},
// 						})
// 					}
// 				}
// 			}

// 			for _, attemptDetails := range runner.resourceTracker.AttemptToDeleteAllAsYetUndeletedResources() {
// 				if attemptDetails.Error != nil {
// 					eventHandler.sayThatResourceDeletionFailed(attemptDetails.Resource.information, attemptDetails.Error, testUnit, testCase)
// 					return
// 				}

// 				eventHandler.sayThatResourceDeletionSucceeded(attemptDetails.Resource.information, testUnit, testCase)
// 			}

// 			eventHandler.sayThatCaseCompletedSuccessfully(testUnit, testCase)
// 		}

// 		eventHandler.sayThatUnitCompletedSuccessfully(testUnit)
// 	}

// 	if err := assetsDirectoryManager.GenerateArchiveFileAt(runner.config.Test.Definition.ArchiveFilePath); err != nil {
// 		eventHandler.sayThatArchiveCreationFailed(runner.config.Test.Definition.ArchiveFilePath, assetsDirectoryManager.TestRootAssetDirectoryPath(), err)
// 		return
// 	}

// 	eventHandler.sayThatArchiveCreationSucceeded(assetsDirectoryManager.TestRootAssetDirectoryPath())

// 	if err := assetsDirectoryManager.RemoveAssetsDirectory(); err != nil {
// 		eventHandler.sayThatAssetDirectoryDeletionFailed(assetsDirectoryManager.TestRootAssetDirectoryPath(), err)
// 		return
// 	}

// 	eventHandler.sayThatAssetDirectoryDeletionWasSuccessful(assetsDirectoryManager.TestRootAssetDirectoryPath())

// 	eventHandler.sayThatTestingCompletedSuccessfully()
// }

// func writeActionOutcomeInformationToAssetsDirectory(action *PipelineAction, outcome *PipelineActionOutcome, testCaseAssetsDirectories *TestCaseDirectoryPaths) {
// 	switch action.Type {
// 	case TemplatedResource:
// 		outputFilesBasePath := deriveActionOutputFilesBasePath(testCaseAssetsDirectories.ExpandedTemplates, action.ActionFullyQualifiedPath)
// 		outcome.WriteOutputToFile(outputFilesBasePath, 0600)

// 	case Executable:
// 		outputFilesBasePath := deriveActionOutputFilesBasePath(testCaseAssetsDirectories.Executables, action.ActionFullyQualifiedPath)
// 		outcome.WriteOutputToFile(fmt.Sprintf("%s.stdout", outputFilesBasePath), 0600)
// 		outcome.WriteErrorToFile(fmt.Sprintf("%s.stderr", outputFilesBasePath), 0600)

// 	case ValuesTransform:
// 		outputFilesBasePath := deriveActionOutputFilesBasePath(testCaseAssetsDirectories.ValuesTransforms, action.ActionFullyQualifiedPath)
// 		outcome.WriteOutputToFile(fmt.Sprintf("%s.stdout", outputFilesBasePath), 0600)
// 		outcome.WriteErrorToFile(fmt.Sprintf("%s.stderr", outputFilesBasePath), 0600)
// 	}
// }

// func deriveActionOutputFilesBasePath(depositDirectoryPath string, actionFullyQualifiedName string) string {
// 	actionFullyQalifiedNamePathElements := strings.Split(actionFullyQualifiedName, "/")
// 	actionBasename := actionFullyQalifiedNamePathElements[len(actionFullyQalifiedNamePathElements)-1]

// 	basePath := fmt.Sprintf("%s/%s", depositDirectoryPath, actionBasename)
// 	candidatePath := basePath

// 	for discriminatorInt := 0; fileExists(candidatePath); discriminatorInt++ {
// 		candidatePath = fmt.Sprintf("%s-%d", basePath, discriminatorInt)
// 	}

// 	return candidatePath
// }

// func fileExists(filePath string) bool {
// 	_, err := os.Stat(filePath)

// 	switch {
// 	case err == nil:
// 		return true
// 	case errors.Is(err, os.ErrNotExist):
// 		return false
// 	default:
// 		panic(fmt.Sprintf("os.Stat failed: %s", err))
// 	}
// }
