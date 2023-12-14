package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/blorticus-go/jobber"
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

func (l *Logger) SayContextually(context jobber.EventContext, formatString string, a ...any) {
	f := fmt.Sprintf("%s %s", l.possiblyFixedWidthContextString(context), formatString)
	l.Say(f, a...)
}

func (l *Logger) possiblyFixedWidthContextString(context jobber.EventContext) string {
	if context.CaseName == "" {
		r := fmt.Sprintf("[%s]", context.UnitName)
		return fmt.Sprintf(l.contextFieldWidthPrintfSpecifier, r)
	}

	r := fmt.Sprintf("[%s/%s]", context.UnitName, context.CaseName)
	return fmt.Sprintf(l.contextFieldWidthPrintfSpecifier, r)
}

func (l *Logger) LogEventMessage(event *jobber.Event) {
	switch event.Type {
	case jobber.ResourceCreationSuccess:
		l.SayContextually(event.Context, "Successfully created resource kind [%s] named [%s]", event.ResourceInformation.ResourceDetails.Kind, event.ResourceInformation.ResourceDetails.Name)
	case jobber.ResourceCreationFailure:
		l.SayContextually(event.Context, "Failed to create resource kind [%s] named [%s]: %s\nTemplate:\n%s\n", event.ResourceInformation.ResourceDetails.Kind, event.ResourceInformation.ResourceDetails.Name, event.Error, event.ResourceInformation.ExpandedTemplateRetriever())
	case jobber.ResourceTemplateExpansionFailure:
		l.SayContextually(event.Context, "Template expansion failure: %s", event.Error)
	case jobber.ResourceDeletionSuccess:
		l.SayContextually(event.Context, "Successfully deleted resource kind [%s] named [%s]", event.ResourceInformation.ResourceDetails.Kind, event.ResourceInformation.ResourceDetails.Name)
	case jobber.ResourceDeletionFailure:
		l.SayContextually(event.Context, "Failed to delete resource kind [%s] named [%s]: %s", event.ResourceInformation.ResourceDetails.Kind, event.ResourceInformation.ResourceDetails.Name, event.Error)
	case jobber.ValuesTransformSuccess:
		l.SayContextually(event.Context, "ValueTransform [%s] completed successfully", event.ValuesTransformInformation.TransformerName)
	case jobber.ValuesTransformFailure:
		l.SayContextually(event.Context, "ValueTransform [%s] failed: %s", event.ValuesTransformInformation.TransformerName, event.Error)
	case jobber.ExecutableRunSuccess:
		l.SayContextually(event.Context, "Executable [%s] ran successfully", event.ExecuableInformation.ExecutableName)
	case jobber.ExecutableRunFailure:
		l.SayContextually(event.Context, "Executable [%s] run failed: %s", event.ExecuableInformation.ExecutableName, event.Error)
	case jobber.TestUnitStarted:
		l.SayContextually(event.Context, "Unit started")
	case jobber.TestUnitCompletedSuccessfully:
		l.SayContextually(event.Context, "Unit completed succesfully")
	case jobber.TestCaseStarted:
		l.SayContextually(event.Context, "Test case started")
	case jobber.TestCaseCompletedSuccessfully:
		l.SayContextually(event.Context, "Test case completed succesfully")
	case jobber.TestingCompletedSuccesfully:
		l.SayContextually(event.Context, "Testing completed successfully")
	case jobber.AssetDirectoryCreatedSuccessfully:
		l.SayContextually(event.Context, "Created directory [%s]", event.FileEvent.Path)
	case jobber.AssetDirectoryCreationFailed:
		if event.FileEvent.Path == "" {
			if event.Context.UnitName == "" {
				l.SayContextually(event.Context, "Failed to create asset root temp directory: %s", event.Error)
			} else {
				l.SayContextually(event.Context, "Failed to create temp directory: %s", event.Error)
			}
		} else {
			l.SayContextually(event.Context, "Failed to create directory [%s]: %s", event.FileEvent.Path, event.Error)
		}
	}
}

func justTheTargetFrom(typeAndTarget string) string {
	return (strings.Split(typeAndTarget, "/"))[1]
}
