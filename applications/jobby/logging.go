package main

import (
	"fmt"
	"io"
	"os"
)

type Logger struct {
	fatalMessageDestination  io.Writer
	normalMessageDestination io.Writer
}

func NewLogger() *Logger {
	return &Logger{
		fatalMessageDestination:  os.Stderr,
		normalMessageDestination: os.Stdout,
	}
}

func (l *Logger) fatalFinalEvent() {
	os.Exit(1)
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
