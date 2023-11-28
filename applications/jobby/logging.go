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

func (l *Logger) DieIfError(err error, formatString string, a ...any) {
	if err != nil {
		if formatString != "" {
			fmt.Fprintf(l.fatalMessageDestination, formatString+": ", a...)
		}
		fmt.Printf("%s\n", err.Error())
		l.fatalFinalEvent()
	}
}

func (l *Logger) Say(formatString string, a ...any) {
	fmt.Fprintf(l.normalMessageDestination, formatString+"\n", a...)
}
