package jobber

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

func NewEventContext(testUnit *TestUnit, testCase *TestCase) EventContext {
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
	ResourceInformation *K8sResourceInformation

	// Error will be non-nil if the event is ResourceCreatingFailure or ResourceDeletionFailure.
	Error error
}

type ValuesTransformEvent struct {
	InputValuesRetriever  StringRetriever
	OutputValuesRetriever StringRetriever
	StderrOutputRetriever StringRetriever
	Error                 error
}

type ExecutableEvent struct {
	StdoutOutputRetriever StringRetriever
	StderrOutputRetriever StringRetriever
	Error                 error
}

type Event struct {
	Type                       EventType
	PipelinePathId             string
	Context                    EventContext
	ResourceInformation        *ResourceEvent
	ValuesTransformInformation *ValuesTransformEvent
	ExecuableInformation       *ExecutableEvent
	ForcedTestToAbortOnError   bool
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

func (h *eventHandler) sayThatResourceCreationFailed(resourceInformation *K8sResourceInformation, pipelinePathId string, templateRetrieverMethod StringRetriever, err error, testUnit *TestUnit, testCase *TestCase) {
	h.eventChannel <- &Event{
		Type: ResourceCreationFailure,
		ResourceInformation: &ResourceEvent{
			ExpandedTemplateRetriever: templateRetrieverMethod,
			ResourceInformation:       resourceInformation,
			Error:                     err,
		},
		Context:                  NewEventContext(testUnit, testCase),
		PipelinePathId:           pipelinePathId,
		ForcedTestToAbortOnError: true,
	}
}

func (h *eventHandler) sayThatResourceCreationSucceeded(resourceInformation *K8sResourceInformation, pipelinePathId string, templateRetrieverMethod StringRetriever, testUnit *TestUnit, testCase *TestCase) {
	h.eventChannel <- &Event{
		Type: ResourceCreationSuccess,
		ResourceInformation: &ResourceEvent{
			ExpandedTemplateRetriever: templateRetrieverMethod,
			ResourceInformation:       resourceInformation,
		},
		Context:        NewEventContext(testUnit, testCase),
		PipelinePathId: pipelinePathId,
	}
}
