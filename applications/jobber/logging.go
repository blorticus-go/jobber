package main

import (
	"fmt"
	"io"
	"os"

	"github.com/blorticus-go/jobber"
	"github.com/blorticus-go/jobber/wrapped"
)

type Logger struct {
	fatalMessageDestination          io.Writer
	normalMessageDestination         io.Writer
	contextFieldWidthPrintfSpecifier string
}

func NewLogger() *Logger {
	return &Logger{
		fatalMessageDestination:          os.Stderr,
		normalMessageDestination:         os.Stdout,
		contextFieldWidthPrintfSpecifier: "%s",
	}
}

type LoggedMessageType string

const (
	Started      LoggedMessageType = "Started"
	TryingTo     LoggedMessageType = "Trying To"
	Successfully LoggedMessageType = "Successfully"
	FailedTo     LoggedMessageType = "Failed To"
	WaitingFor   LoggedMessageType = "Waiting For"
)

var charactersInLongestLoggedMessageType = len(Successfully)

func (l *Logger) fatalFinalEvent() {
	os.Exit(1)
}

func (l *Logger) SetContextFieldWidth(maximumUnitNameLength uint, maximumCaseNameLength uint) *Logger {
	fieldWidth := maximumUnitNameLength + maximumCaseNameLength + 3 // +1 for delimiter and brackets
	l.contextFieldWidthPrintfSpecifier = fmt.Sprintf("%%-%d.%ds", fieldWidth, fieldWidth)
	return l
}

func (l *Logger) Fatalf(formatString string, a ...any) {
	fmt.Fprintf(l.fatalMessageDestination, formatString, a...)
	l.fatalFinalEvent()
}

func (l *Logger) DieIfError(err error, formatStringThenSprintfArgs ...any) {
	if err != nil {
		if len(formatStringThenSprintfArgs) != 0 {
			fmt.Fprintf(l.fatalMessageDestination, formatStringThenSprintfArgs[0].(string)+": ", formatStringThenSprintfArgs[1:]...)
		}
		fmt.Printf("%s\n", err.Error())
		l.fatalFinalEvent()
	}
}

func (l *Logger) Say(formatString string, a ...any) {
	fmt.Fprintf(l.normalMessageDestination, formatString+"\n", a...)
}

func (l *Logger) SayContextually(context jobber.EventContext, messageType LoggedMessageType, formatString string, a ...any) {
	f := fmt.Sprintf("%s  %s  %s", l.possiblyFixedWidthContextString(context), l.eventTypeString(messageType), formatString)
	l.Say(f, a...)
}

func (l *Logger) eventTypeString(messageType LoggedMessageType) string {
	formatString := fmt.Sprintf("%%-%d.%ds", charactersInLongestLoggedMessageType, charactersInLongestLoggedMessageType)
	return fmt.Sprintf(formatString, string(messageType))
}

func (l *Logger) possiblyFixedWidthContextString(context jobber.EventContext) string {
	if context.TestCaseName == "" {
		r := fmt.Sprintf("[%s]", context.TestUnitName)
		return fmt.Sprintf(l.contextFieldWidthPrintfSpecifier, r)
	}

	r := fmt.Sprintf("[%s/%s]", context.TestUnitName, context.TestCaseName)
	return fmt.Sprintf(l.contextFieldWidthPrintfSpecifier, r)
}

func resourceDescriptionFor(resource wrapped.Resource) string {
	if resource.NamespaceName() == "" {
		return fmt.Sprintf("Kind [%s] Named [%s]", resource.GroupVersionKind().Kind, resource.Name())
	}

	return fmt.Sprintf("Kind [%s] Named [%s] in Namespace [%s]", resource.GroupVersionKind().Kind, resource.Name(), resource.NamespaceName())
}

func (l *Logger) LogEventMessage(event *jobber.Event) {
	switch event.Type {
	case jobber.UnitStarted:
		l.SayContextually(event.Context, Started, "Unit")
	case jobber.CaseStarted:
		l.SayContextually(event.Context, Started, "Case")
	case jobber.UnitCompletedSuccessfully:
		l.SayContextually(event.Context, Successfully, "Completed Unit")
	case jobber.CaseCompletedSuccessfully:
		l.SayContextually(event.Context, Successfully, "Completed Case")
	case jobber.TestCompletedSuccessfully:
		l.SayContextually(event.Context, Successfully, "Completed Test")
	case jobber.TryingToCreateResource:
		if event.Resource.Name() == "" && event.Resource.GroupVersionKind().Kind == "Namespace" {
			l.SayContextually(event.Context, TryingTo, "Create default Namespace")
		} else {
			l.SayContextually(event.Context, TryingTo, "Create Resource %s", resourceDescriptionFor(event.Resource))
		}
	case jobber.SuccessfullyCreatedResource:
		l.SayContextually(event.Context, Successfully, "Created Resource %s", resourceDescriptionFor(event.Resource))
	case jobber.FailedToCreateResource:
		l.SayContextually(event.Context, FailedTo, "Create Resource %s: %s", resourceDescriptionFor(event.Resource), event.Error)
	case jobber.SuccessfullyDeletedResource:
		l.SayContextually(event.Context, Successfully, "Deleted Resource %s", resourceDescriptionFor(event.Resource))
	case jobber.FailedToDeleteResource:
		l.SayContextually(event.Context, FailedTo, "Delete Resource %s: %s", resourceDescriptionFor(event.Resource), event.Error)
	case jobber.WaitingForJobToComplete:
		l.SayContextually(event.Context, WaitingFor, "Job Named [%s] in Namespace [%s] to Complete", event.Resource.Name(), event.Resource.NamespaceName())
	case jobber.JobCompletedSuccessfully:
		l.SayContextually(event.Context, Successfully, "Completed Job Named [%s] in Namespace [%s]", resourceDescriptionFor(event.Resource), event.Resource.NamespaceName())
	case jobber.JobFailedToCompleteSuccessfully:
		l.SayContextually(event.Context, FailedTo, "Complete Job Named [%s] in Namespace [%s]: %s", resourceDescriptionFor(event.Resource), event.Resource.NamespaceName(), event.Error)
	case jobber.SuccessfullyCreatedFile:
		l.SayContextually(event.Context, Successfully, "Created Directory [%s]", event.FileOrDirectoryPath)
	case jobber.SuccessfullyCreatedDirectory:
		l.SayContextually(event.Context, Successfully, "Created File [%s]", event.FileOrDirectoryPath)
	case jobber.FailedToCreateFile:
		l.SayContextually(event.Context, FailedTo, "Create File [%s]: %s", event.FileOrDirectoryPath, event.Error)
	case jobber.FailedToCreateDirectory:
		l.SayContextually(event.Context, FailedTo, "Create Directory [%s]: %s", event.FileOrDirectoryPath, event.Error)
	case jobber.FailedToProcessPipelineDescriptors:
		l.SayContextually(event.Context, FailedTo, "Process Pipeline Action Descriptors: %s", event.Error)
	case jobber.FailedToProcessTemplate:
		l.SayContextually(event.Context, FailedTo, "Process A Template: %s", event.Error)
	}
}
