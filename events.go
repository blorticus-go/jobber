package jobber

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

type StringRetriever func() string

type K8sResourceInformation struct {
	Kind          string
	Name          string
	NamespaceName string
}

type EventContext struct {
	CaseName string
	UnitName string
}

func EventContextFor(testUnit *TestUnit, testCase *TestCase) EventContext {
	if testUnit == nil {
		return EventContext{}
	}

	if testCase == nil {
		return EventContext{
			UnitName: testUnit.Name,
		}
	}

	return EventContext{
		UnitName: testUnit.Name,
		CaseName: testCase.Name,
	}
}

type EventType int

const (
	ResourceCreationSuccess EventType = iota
	ResourceCreationFailure
	ResourceTemplateExpansionFailure
	ResourceDeletionSuccess
	ResourceDeletionFailure
	ValuesTransformSuccess
	ValuesTransformFailure
	ExecutableRunSuccess
	ExecutableRunFailure
	TestUnitStarted
	TestUnitCompletedSuccessfully
	TestCaseStarted
	TestCaseCompletedSuccessfully
	TestingCompletedSuccesfully
	PipelineDefinitionIsInvalid
	AssetDirectoryCreatedSuccessfully
	AssetDirectoryCreationFailed
	WaitingForPodToReachRunningState
	WaitingForJobToComplete
	JobFailedToComplete
)

type ResourceEvent struct {
	// ExpandedTemplateRetriever provides a StringRetriever method that generates the template (as a string) after it has been
	// expanded (i.e., the go-template transforms have run).  If the Resource is built-in (e.g., the Default
	// Namespace) this will be nil.
	ExpandedTemplateRetriever StringRetriever

	// ResourceInformation describes the Kubernetes Resource to which the event pertains.
	// If there was an error and the template didn't provide enough information to determine
	// all of the information for the Resource, this value will be nil.
	ResourceDetails *K8sResourceInformation

	// The pipeline action identifier for the template.  This is set only when the event type is ResourceTemplateExpansionFailure.
	TemplateName string
}

type ValuesTransformEvent struct {
	TransformerName       string
	InputValuesRetriever  StringRetriever
	OutputValuesRetriever StringRetriever
	StderrOutputRetriever StringRetriever
}

type ExecutableEvent struct {
	ExecutableName        string
	StdoutOutputRetriever StringRetriever
	StderrOutputRetriever StringRetriever
}

type FileEvent struct {
	Path string
}

type Event struct {
	Type                       EventType
	Context                    EventContext
	ResourceInformation        *ResourceEvent
	ValuesTransformInformation *ValuesTransformEvent
	ExecuableInformation       *ExecutableEvent
	FileEvent                  *FileEvent
	Error                      error
}

type eventHandler struct {
	eventChannel chan<- *Event
}

func (h *eventHandler) sayThatUnitStarted(testUnit *TestUnit) {
	h.eventChannel <- &Event{
		Type: TestUnitStarted,
		Context: EventContext{
			UnitName: testUnit.Name,
		},
	}
}

func (h *eventHandler) sayThatUnitCompletedSuccessfully(testUnit *TestUnit) {
	h.eventChannel <- &Event{
		Type: TestUnitCompletedSuccessfully,
		Context: EventContext{
			UnitName: testUnit.Name,
		},
	}
}

func (h *eventHandler) sayThatCaseStarted(testUnit *TestUnit, testCase *TestCase) {
	h.eventChannel <- &Event{
		Type: TestCaseStarted,
		Context: EventContext{
			UnitName: testUnit.Name,
			CaseName: testCase.Name,
		},
	}
}

func (h *eventHandler) sayThatCaseCompletedSuccessfully(testUnit *TestUnit, testCase *TestCase) {
	h.eventChannel <- &Event{
		Type: TestCaseCompletedSuccessfully,
		Context: EventContext{
			UnitName: testUnit.Name,
			CaseName: testCase.Name,
		},
	}
}

func (h *eventHandler) sayThatTestingCompletedSuccessfully() {
	h.eventChannel <- &Event{
		Type: TestingCompletedSuccesfully,
	}
}

func (h *eventHandler) sayThatResourceCreationSucceeded(resourceInformation *K8sResourceInformation, templateRetrieverMethod StringRetriever, testUnit *TestUnit, testCase *TestCase) {
	h.eventChannel <- &Event{
		Type: ResourceCreationSuccess,
		ResourceInformation: &ResourceEvent{
			ExpandedTemplateRetriever: templateRetrieverMethod,
			ResourceDetails:           resourceInformation,
		},
		Context: EventContextFor(testUnit, testCase),
	}
}

func (h *eventHandler) sayThatResourceCreationFailed(resourceInformation *K8sResourceInformation, templateRetrieverMethod StringRetriever, err error, testUnit *TestUnit, testCase *TestCase) {
	h.eventChannel <- &Event{
		Type: ResourceCreationFailure,
		ResourceInformation: &ResourceEvent{
			ExpandedTemplateRetriever: templateRetrieverMethod,
			ResourceDetails:           resourceInformation,
		},
		Context: EventContextFor(testUnit, testCase),
		Error:   err,
	}
}

func (h *eventHandler) sayThatResourceTemplateExpansionFailed(templateName string, templateRetrieverMethod StringRetriever, err error, testUnit *TestUnit, testCase *TestCase) {
	h.eventChannel <- &Event{
		Type: ResourceTemplateExpansionFailure,
		ResourceInformation: &ResourceEvent{
			ExpandedTemplateRetriever: templateRetrieverMethod,
			TemplateName:              templateName,
		},
		Context: EventContextFor(testUnit, testCase),
		Error:   fmt.Errorf("on template (%s) %s", templateName, err),
	}
}

func (h *eventHandler) explainAttemptToCreateDefaultNamespace(generatedNameBase string, createdNamespaceApiObject *corev1.Namespace, context EventContext, errorOnCreationAttempt error) {
	var namespaceName string
	if createdNamespaceApiObject != nil {
		namespaceName = createdNamespaceApiObject.Name
	} else {
		namespaceName = fmt.Sprintf("%s-<generated>", generatedNameBase)
	}

	if errorOnCreationAttempt != nil {
		h.eventChannel <- &Event{
			Type:    ResourceCreationFailure,
			Context: context,
			ResourceInformation: &ResourceEvent{
				ResourceDetails: &K8sResourceInformation{
					Kind:          "namespace",
					Name:          namespaceName,
					NamespaceName: "",
				},
			},
			Error: errorOnCreationAttempt,
		}
	} else {
		h.eventChannel <- &Event{
			Type:    ResourceCreationSuccess,
			Context: context,
			ResourceInformation: &ResourceEvent{
				ResourceDetails: &K8sResourceInformation{
					Kind:          "namespace",
					Name:          namespaceName,
					NamespaceName: "",
				},
			},
		}
	}
}

func (h *eventHandler) sayThatPipelineDefinitionIsInvalid(err error) {
	h.eventChannel <- &Event{
		Type:    PipelineDefinitionIsInvalid,
		Context: EventContext{},
		Error:   err,
	}
}

func (h *eventHandler) explainActionOutcome(action *PipelineAction, outcome *PipelineActionOutcome, testUnit *TestUnit, testCase *TestCase) {
	switch action.Type {
	case TemplatedResource:
		templateOutputFunc := func() string {
			if outcome.OutputBuffer != nil {
				return outcome.OutputBuffer.String()
			}

			return ""
		}

		if outcome.Error == nil {
			h.sayThatResourceCreationSucceeded(outcome.CreatedResource.Information(), templateOutputFunc, testUnit, testCase)
		} else {
			switch typedError := outcome.Error.(type) {
			case *TemplateError:
				h.sayThatResourceTemplateExpansionFailed(typedError.TemplateName, templateOutputFunc, typedError, testUnit, testCase)
			case *ResourceCreationError:
				h.sayThatResourceCreationFailed(typedError.ResourceInformation, templateOutputFunc, typedError, testUnit, testCase)
			case *JobCompletionFailureError:
				h.sayThatJobFailedToComplete(typedError.ResourceInformation, typedError, testUnit, testCase)
			}
		}

	case ValuesTransform:
	case Executable:
	}
}

func (handler *eventHandler) sayThatAssetDirectoryCreationFailed(path string, err error, testUnit *TestUnit, testCase *TestCase) {
	handler.eventChannel <- &Event{
		Type: AssetDirectoryCreationFailed,
		FileEvent: &FileEvent{
			Path: path,
		},
		Error:   err,
		Context: EventContextFor(testUnit, testCase),
	}
}

func (handler *eventHandler) sayThatJobFailedToComplete(resourceInformation *K8sResourceInformation, err error, testUnit *TestUnit, testCase *TestCase) {
	handler.eventChannel <- &Event{
		Type:    JobFailedToComplete,
		Error:   err,
		Context: EventContextFor(testUnit, testCase),
	}
}

func (handler *eventHandler) sayThatAssetDirectoryCreationSucceeded(directoryPath string, testUnit *TestUnit, testCase *TestCase) {
	handler.eventChannel <- &Event{
		Type: AssetDirectoryCreatedSuccessfully,
		FileEvent: &FileEvent{
			Path: directoryPath,
		},
		Context: EventContextFor(testUnit, testCase),
	}
}

func (handler *eventHandler) explainAssetCreationOutcome(outcome *TestCaseAssetsDirectoryCreationOutcome, testUnit *TestUnit, testCase *TestCase) {
	for _, successfullyCreatedDir := range outcome.SuccessfullyCreatedDirectoryPaths {
		handler.sayThatAssetDirectoryCreationSucceeded(successfullyCreatedDir, testUnit, testCase)
	}
	if outcome.DirectoryCreationFailureError != nil {
		handler.sayThatAssetDirectoryCreationFailed(outcome.DirectoryPathOfFailedCreation, outcome.DirectoryCreationFailureError, testUnit, testCase)
	}
}
