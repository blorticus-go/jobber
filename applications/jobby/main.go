package main

import (
	"os"
	"text/template"
)

var template01 string = `---
apiVersion: v1
kind: {{ .Values.Kind }}
`

type T struct {
	Values map[string]any
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

type CommandLineArguments struct {
	ConfigurationFilePath string
}

func (args *CommandLineArguments) fillDefaults() *CommandLineArguments {
	return args
}

func (args *CommandLineArguments) validate() error {
	return nil
}

func main() {
	tmpl, err := template.New("test").Parse(template01)
	panicIfError(err)

	t := T{Values: map[string]any{"Kind": 10}}

	err = tmpl.Execute(os.Stdout, t)
	panicIfError(err)
}
