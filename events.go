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
	TryingToCreateDirectory
	TryingToCreateFile
	SuccessfullyCreatedFile
	SuccessfullyCreatedDirectory
	FailedToCreateFile
	FailedToCreateDirectory
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

type ScopedEventFactory struct {
	currentContext EventContext
}

func NewGlobalScopedEventFactory() *ScopedEventFactory {
	return &ScopedEventFactory{
		currentContext: EventContext{},
	}
}

func (factory *ScopedEventFactory) ScopedToUnitNamed(testUnitName string) *ScopedEventFactory {
	return &ScopedEventFactory{
		currentContext: EventContext{
			TestUnitName: testUnitName,
		},
	}
}

func (factory *ScopedEventFactory) ScopedToCaseNamed(testCaseName string) *ScopedEventFactory {
	return &ScopedEventFactory{
		currentContext: EventContext{
			TestUnitName: factory.currentContext.TestUnitName,
			TestCaseName: testCaseName,
		},
	}
}

func (factory *ScopedEventFactory) NewUnitStartedEvent() *Event {
	return &Event{
		Type:    UnitStarted,
		Context: factory.currentContext,
	}
}

func (factory *ScopedEventFactory) NewCaseStartedEvent() *Event {
	return &Event{
		Type:    UnitStarted,
		Context: factory.currentContext,
	}
}

func (factory *ScopedEventFactory) NewUnitCompletedSuccessfullyEvent() *Event {
	return &Event{
		Type:    UnitCompletedSuccessfully,
		Context: factory.currentContext,
	}
}

func (factory *ScopedEventFactory) NewCaseCompletedSuccessfullyEvent() *Event {
	return &Event{
		Type:    CaseCompletedSuccessfully,
		Context: factory.currentContext,
	}
}

func (factory *ScopedEventFactory) NewTestCompletedSuccessfullyEvent() *Event {
	return &Event{
		Type:    TestCompletedSuccessfully,
		Context: factory.currentContext,
	}
}

func (factory *ScopedEventFactory) NewTryingToCreateResourceEvent(resource wrapped.Resource) *Event {
	return &Event{
		Type:     TryingToCreateResource,
		Resource: resource,
		Context:  factory.currentContext,
	}
}

func (factory *ScopedEventFactory) NewSuccessfullyCreatedResourceEvent(resource wrapped.Resource) *Event {
	return &Event{
		Type:     SuccessfullyCreatedResource,
		Resource: resource,
		Context:  factory.currentContext,
	}
}

func (factory *ScopedEventFactory) NewFailedToCreateResourceEvent(resource wrapped.Resource, err error) *Event {
	return &Event{
		Type:     FailedToCreateResource,
		Resource: resource,
		Context:  factory.currentContext,
		Error:    err,
	}
}

func (factory *ScopedEventFactory) NewSuccessfullyDeletedResourceEvent(resource wrapped.Resource) *Event {
	return &Event{
		Type:     SuccessfullyDeletedResource,
		Resource: resource,
		Context:  factory.currentContext,
	}
}

func (factory *ScopedEventFactory) NewFailedToDeleteResourceEvent(resource wrapped.Resource, err error) *Event {
	return &Event{
		Type:     FailedToDeleteResource,
		Resource: resource,
		Context:  factory.currentContext,
		Error:    err,
	}
}

func (factory *ScopedEventFactory) NewWaitingForJobToCompleteEvent(resource wrapped.Resource) *Event {
	return &Event{
		Type:    TestCompletedSuccessfully,
		Context: factory.currentContext,
	}
}

func (factory *ScopedEventFactory) NewJobCompletedSuccessfullyEvent(resource wrapped.Resource) *Event {
	return &Event{
		Type:    TestCompletedSuccessfully,
		Context: factory.currentContext,
	}
}

func (factory *ScopedEventFactory) NewJobFailedToCompleteSuccessfullyEvent(resource wrapped.Resource, err error) *Event {
	return &Event{
		Type:    TestCompletedSuccessfully,
		Context: factory.currentContext,
		Error:   err,
	}
}

func (factory *ScopedEventFactory) NewTryingToCreateDirectoryEvent(directoryPath string) *Event {
	return &Event{
		Type:                TestCompletedSuccessfully,
		Context:             factory.currentContext,
		FileOrDirectoryPath: directoryPath,
	}
}

func (factory *ScopedEventFactory) NewTryingToCreateFileEvent(filePath string) *Event {
	return &Event{
		Type:                TestCompletedSuccessfully,
		Context:             factory.currentContext,
		FileOrDirectoryPath: filePath,
	}
}

func (factory *ScopedEventFactory) NewSuccessfullyCreatedFileEvent(filePath string) *Event {
	return &Event{
		Type:                TestCompletedSuccessfully,
		Context:             factory.currentContext,
		FileOrDirectoryPath: filePath,
	}
}

func (factory *ScopedEventFactory) NewSuccessfullyCreatedDirectoryEvent(directoryPath string) *Event {
	return &Event{
		Type:                TestCompletedSuccessfully,
		Context:             factory.currentContext,
		FileOrDirectoryPath: directoryPath,
	}
}

func (factory *ScopedEventFactory) NewFailedToCreateFileEvent(filePath string, err error) *Event {
	return &Event{
		Type:                TestCompletedSuccessfully,
		Context:             factory.currentContext,
		FileOrDirectoryPath: filePath,
		Error:               err,
	}
}

func (factory *ScopedEventFactory) NewFailedToCreateDirectoryEvent(directoryPath string, err error) *Event {
	return &Event{
		Type:                TestCompletedSuccessfully,
		Context:             factory.currentContext,
		FileOrDirectoryPath: directoryPath,
		Error:               err,
	}
}
