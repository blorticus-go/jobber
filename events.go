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

type Event struct {
	Type                       EventType
	Context                    EventContext
	ResourceInformation        *ResourceEvent
	ValuesTransformInformation *ValuesTransformEvent
	ExecuableInformation       *ExecutableEvent
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

func (h *eventHandler) sayThatResourceDeletionSucceeded(resourceInformation *K8sResourceInformation, testUnit *TestUnit, testCase *TestCase) {
	h.eventChannel <- &Event{
		Type: ResourceDeletionSuccess,
		ResourceInformation: &ResourceEvent{
			ResourceDetails: resourceInformation,
		},
		Context: EventContextFor(testUnit, testCase),
	}
}

func (h *eventHandler) sayThatResourceDeletionFailed(resourceInformation *K8sResourceInformation, err error, testUnit *TestUnit, testCase *TestCase) {
	h.eventChannel <- &Event{
		Type: ResourceDeletionFailure,
		ResourceInformation: &ResourceEvent{
			ResourceDetails: resourceInformation,
		},
		Context: EventContextFor(testUnit, testCase),
		Error:   err,
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
