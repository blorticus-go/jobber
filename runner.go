package jobber

import (
	"context"

	"github.com/qdm12/reprint"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TemplateExpansionNamespace struct {
	GeneratedName string
}

type TemplateExpansionConfigVariables struct {
	Namespaces map[string]*TemplateExpansionNamespace
}

type PipelineVariables struct {
	Values map[string]any
	Config *TemplateExpansionConfigVariables
}

type RunningContext struct {
	testUnit *TestUnit
	testCase *TestCase
}

type Runner struct {
	client               *Client
	config               *Configuration
	defaultNamespaceName string
	resourceTracker      *CreatedResourceTracker
}

func NewRunner(config *Configuration, client *Client) *Runner {
	return &Runner{
		client:          client,
		config:          config,
		resourceTracker: NewCreatedResourceTracker(),
	}
}

func (runner *Runner) createDefaultNamespace() (*corev1.Namespace, error) {
	defaultNamespaceApiObject, err := runner.client.CreateNamespaceUsingGeneratedName(runner.config.Test.Definition.Namespaces["Default"].Basename)
	if err != nil {
		return nil, err
	}

	runner.resourceTracker.AddCreatedResource(&K8sResource{
		information: &K8sResourceInformation{
			Kind:          "namespace",
			Name:          defaultNamespaceApiObject.Name,
			NamespaceName: "",
		},
		deletionMethod: func(object any) error {
			return runner.client.clientSet.CoreV1().Namespaces().Delete(context.Background(), defaultNamespaceApiObject.Name, metav1.DeleteOptions{})
		},
	})

	return defaultNamespaceApiObject, nil
}

func (runner *Runner) RunTest(eventChannel chan<- *Event) {
	eventHandler := &eventHandler{eventChannel}
	assetsDirectoryManager := NewContextualAssetsDirectoryManager()

	if err := assetsDirectoryManager.CreateTestAssetsRootDirectory(); err != nil {
		eventHandler.sayThatAssetDirectoryCreationFailed(assetsDirectoryManager.TestRootAssetDirectoryPath(), err, nil, nil)
		return
	} else {
		eventHandler.sayThatAssetDirectoryCreationSucceeded(assetsDirectoryManager.TestRootAssetDirectoryPath(), nil, nil)
	}

	testCasePipeline, err := NewPipelineFromStringDescriptors(runner.config.Test.Definition.Pipeline, runner.config.Test.Definition.PipelineRootDirectory)
	if err != nil {
		eventHandler.sayThatPipelineDefinitionIsInvalid(err)
		return
	}

	for _, testUnit := range runner.config.Test.Units {
		eventHandler.sayThatUnitStarted(testUnit)

		if err := assetsDirectoryManager.CreateTestUnitDirectory(testUnit); err != nil {
			eventHandler.sayThatAssetDirectoryCreationFailed(assetsDirectoryManager.TestRootAssetDirectoryPath(), err, testUnit, nil)
			return
		} else {
			eventHandler.sayThatAssetDirectoryCreationSucceeded(assetsDirectoryManager.TestUnitAssetDirectoryPathFor(testUnit), testUnit, nil)
		}

		for _, testCase := range runner.config.Test.Cases {
			eventHandler.sayThatCaseStarted(testUnit, testCase)

			outcome := assetsDirectoryManager.CreateTestCaseDirectories(testUnit, testCase)
			for _, successfullyCreatedDir := range outcome.SuccessfullyCreatedDirectoryPaths {
				eventHandler.sayThatAssetDirectoryCreationSucceeded(successfullyCreatedDir, testUnit, testCase)
			}
			if outcome.DirectoryCreationFailureError != nil {
				eventHandler.sayThatAssetDirectoryCreationFailed(outcome.DirectoryPathOfFailedCreation, outcome.DirectoryCreationFailureError, testUnit, testCase)
				return
			}

			nsObject, err := runner.createDefaultNamespace()
			eventHandler.explainAttemptToCreateDefaultNamespace(runner.config.Test.Definition.Namespaces["Default"].Basename, nsObject, EventContextFor(testUnit, testCase), err)
			if err != nil {
				return
			}

			testCaseScopedCopyOfVariables := &PipelineVariables{
				Values: reprint.This(testCase.Values).(map[string]any),
				Config: &TemplateExpansionConfigVariables{
					Namespaces: map[string]*TemplateExpansionNamespace{
						"Default": {
							GeneratedName: nsObject.Name,
						},
					},
				},
			}

			for action := testCasePipeline.Restart(); action != nil; action = testCasePipeline.NextAction() {
				for _, outcome := range action.Run(testCaseScopedCopyOfVariables, runner.client) {
					if eventHandler.explainActionOutcome(action, outcome, testUnit, testCase); outcome.Error != nil {
						return
					}
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

	eventHandler.sayThatTestingCompletedSuccessfully()
}
