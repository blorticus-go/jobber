package jobber

import (
	"github.com/blorticus-go/jobber/wrapped"
)

type EventType int

const (
	UnitStarted EventType = iota
	CaseStarted
	UnitCompletedSuccessfully
	CaseCompletedSuccessfully
	TestCompletedSuccessfully
	TryingToCreateResource
	SuccessfullyCreatedResource
	FailedToCreateResource
	SuccessfullyDeletedResource
	FailedToDeleteResource
	WaitingForJobToComplete
	JobCompletedSuccessfully
	JobFailedToCompleteSuccessfully
	SuccessfullyCreatedFile
	SuccessfullyCreatedDirectory
	FailedToCreateFile
	FailedToCreateDirectory
	FailedToProcessPipelineDescriptors
	FailedToProcessTemplate
)

type EventContext struct {
	TestUnitName string
	TestCaseName string
}

func (ec *EventContext) Clone() *EventContext {
	return &EventContext{
		TestUnitName: ec.TestUnitName,
		TestCaseName: ec.TestCaseName,
	}
}

type Event struct {
	Type                EventType
	Context             EventContext
	FileOrDirectoryPath string
	Resource            wrapped.Resource
	Error               error
}

type ScopedEventParrot struct {
	currentContext EventContext
	eventChannel   chan<- *Event
}

func NewGloballyScopedEventParrot(eventChannel chan<- *Event) *ScopedEventParrot {
	return &ScopedEventParrot{
		currentContext: EventContext{},
		eventChannel:   eventChannel,
	}
}

func (parrot *ScopedEventParrot) ScopedToUnitNamed(testUnitName string) *ScopedEventParrot {
	return &ScopedEventParrot{
		currentContext: EventContext{
			TestUnitName: testUnitName,
		},
		eventChannel: parrot.eventChannel,
	}
}

func (parrot *ScopedEventParrot) ScopedToCaseNamed(testCaseName string) *ScopedEventParrot {
	return &ScopedEventParrot{
		currentContext: EventContext{
			TestUnitName: parrot.currentContext.TestUnitName,
			TestCaseName: testCaseName,
		},
		eventChannel: parrot.eventChannel,
	}
}

func (parrot *ScopedEventParrot) SayThatUnitStarted() {
	parrot.eventChannel <- &Event{
		Type:    UnitStarted,
		Context: parrot.currentContext,
	}
}

func (parrot *ScopedEventParrot) SayThatCaseStarted() {
	parrot.eventChannel <- &Event{
		Type:    CaseStarted,
		Context: parrot.currentContext,
	}
}

func (parrot *ScopedEventParrot) SayThatUnitCompletedSuccessfully() {
	parrot.eventChannel <- &Event{
		Type:    UnitCompletedSuccessfully,
		Context: parrot.currentContext,
	}
}

func (parrot *ScopedEventParrot) SayThatCaseCompletedSuccessfully() {
	parrot.eventChannel <- &Event{
		Type:    CaseCompletedSuccessfully,
		Context: parrot.currentContext,
	}
}

func (parrot *ScopedEventParrot) SayThatTestCompletedSuccessfully() {
	parrot.eventChannel <- &Event{
		Type:    TestCompletedSuccessfully,
		Context: parrot.currentContext,
	}
}

func (parrot *ScopedEventParrot) SayThatWeAreTryingToCreateAResource(resource wrapped.Resource) {
	parrot.eventChannel <- &Event{
		Type:     TryingToCreateResource,
		Resource: resource,
		Context:  parrot.currentContext,
	}
}

func (parrot *ScopedEventParrot) SayThatAResourceWasCreatedSuccessfully(resource wrapped.Resource) {
	parrot.eventChannel <- &Event{
		Type:     SuccessfullyCreatedResource,
		Resource: resource,
		Context:  parrot.currentContext,
	}
}

func (parrot *ScopedEventParrot) SayThatWeFailedToCreateAResource(resource wrapped.Resource, err error) {
	parrot.eventChannel <- &Event{
		Type:     FailedToCreateResource,
		Resource: resource,
		Context:  parrot.currentContext,
		Error:    err,
	}
}

func (parrot *ScopedEventParrot) SayThatAResourceWasSuccessfullyDeleted(resource wrapped.Resource) {
	parrot.eventChannel <- &Event{
		Type:     SuccessfullyDeletedResource,
		Resource: resource,
		Context:  parrot.currentContext,
	}
}

func (parrot *ScopedEventParrot) SayThatWeFailedToDeleteAResource(resource wrapped.Resource, err error) {
	parrot.eventChannel <- &Event{
		Type:     FailedToDeleteResource,
		Resource: resource,
		Context:  parrot.currentContext,
		Error:    err,
	}
}

func (parrot *ScopedEventParrot) SayThatWeAreWaitingForAJobToComplete(resource wrapped.Resource) {
	parrot.eventChannel <- &Event{
		Type:     WaitingForJobToComplete,
		Resource: resource,
		Context:  parrot.currentContext,
	}
}

func (parrot *ScopedEventParrot) SayThatAJobSuccessfullyCompleted(resource wrapped.Resource) {
	parrot.eventChannel <- &Event{
		Type:     JobCompletedSuccessfully,
		Resource: resource,
		Context:  parrot.currentContext,
	}
}

func (parrot *ScopedEventParrot) SayThatAJobFailedToComplete(resource wrapped.Resource, err error) {
	parrot.eventChannel <- &Event{
		Type:     JobFailedToCompleteSuccessfully,
		Context:  parrot.currentContext,
		Resource: resource,
		Error:    err,
	}
}

func (parrot *ScopedEventParrot) SayThatAFileWasCreated(filePath string) {
	parrot.eventChannel <- &Event{
		Type:                SuccessfullyCreatedFile,
		Context:             parrot.currentContext,
		FileOrDirectoryPath: filePath,
	}
}

func (parrot *ScopedEventParrot) SayThatADirectoryWasCreated(directoryPath string) {
	parrot.eventChannel <- &Event{
		Type:                SuccessfullyCreatedDirectory,
		Context:             parrot.currentContext,
		FileOrDirectoryPath: directoryPath,
	}
}

func (parrot *ScopedEventParrot) SayThatWeFailedToCreateAFile(filePath string, err error) {
	parrot.eventChannel <- &Event{
		Type:                FailedToCreateFile,
		Context:             parrot.currentContext,
		FileOrDirectoryPath: filePath,
		Error:               err,
	}
}

func (parrot *ScopedEventParrot) SayThatWeFailedToCreateADirectory(directoryPath string, err error) {
	parrot.eventChannel <- &Event{
		Type:                FailedToCreateDirectory,
		Context:             parrot.currentContext,
		FileOrDirectoryPath: directoryPath,
		Error:               err,
	}
}

func (parrot *ScopedEventParrot) SayThatWeCouldNotProcessThePipelineDescriptors(err error) {
	parrot.eventChannel <- &Event{
		Type:    FailedToProcessPipelineDescriptors,
		Context: parrot.currentContext,
		Error:   err,
	}
}

func (parrot *ScopedEventParrot) SayThatWeFailedToProcessATemplate(err error) {
	parrot.eventChannel <- &Event{
		Type:    FailedToProcessTemplate,
		Context: parrot.currentContext,
		Error:   err,
	}
}
