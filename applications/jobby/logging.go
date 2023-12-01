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

func (l *Logger) SayContextually(context *jobber.EventContext, formatString string, a ...any) {
	f := fmt.Sprintf("%s %s", l.possiblyFixedWidthContextString(context), formatString)
	l.Say(f, a...)
}

func (l *Logger) possiblyFixedWidthContextString(context *jobber.EventContext) string {
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
		l.SayContextually(&event.Context, "Resource (%s) created successfully", justTheTargetFrom(event.PipelinePathId))
	case jobber.ResourceCreatingFailure:
		l.SayContextually(&event.Context, "Resource (%s) creation failed: %s", justTheTargetFrom(event.PipelinePathId), event.ResourceInformation.Error)
	case jobber.ResourceDeletionSuccess:
		l.SayContextually(&event.Context, "Resource (%s) deleted successfully", justTheTargetFrom(event.PipelinePathId))
	case jobber.ResourceDeletionFailure:
		l.SayContextually(&event.Context, "Resource (%s) deletion failed: %s", justTheTargetFrom(event.PipelinePathId), event.ResourceInformation.Error)
	case jobber.ValuesTransformSuccess:
		l.SayContextually(&event.Context, "ValueTransform (%s) completed successfully", justTheTargetFrom(event.PipelinePathId))
	case jobber.ValuesTransformFailure:
		l.SayContextually(&event.Context, "ValueTransform (%s) failed: %s", justTheTargetFrom(event.PipelinePathId), event.ValuesTransformInformation.Error)
	case jobber.ExecutableRunSuccess:
		l.SayContextually(&event.Context, "Executable (%s) ran successfully", justTheTargetFrom(event.PipelinePathId))
	case jobber.ExecutableRunFailure:
		l.SayContextually(&event.Context, "Executable (%s) run failed: %s", justTheTargetFrom(event.PipelinePathId), event.ExecuableInformation.Error)
	case jobber.TestUnitStarted:
		l.SayContextually(&event.Context, "Unit started")
	case jobber.TestUnitCompletedSuccessfully:
		l.SayContextually(&event.Context, "Unit completed succesfully")
	case jobber.TestCaseStarted:
		l.SayContextually(&event.Context, "Test case started")
	case jobber.TestCaseCompletedSuccessfully:
		l.SayContextually(&event.Context, "Test case completed succesfully")
	case jobber.TestingCompletedSuccesfully:
		l.SayContextually(&event.Context, "Testing completed successfully")
	}
}

func justTheTargetFrom(typeAndTarget string) string {
	return (strings.Split(typeAndTarget, "/"))[1]
}
