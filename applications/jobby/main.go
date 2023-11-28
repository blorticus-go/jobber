package main

import (
	"fmt"
	"os"

	"github.com/blorticus-go/jobber"
)

var template01 string = `---
apiVersion: v1
kind: {{ .Values.Kind }}
`

type T struct {
	Values map[string]any
}

func dieIfError(err error, f string, a ...interface{}) {
	if err != nil {
		if f != "" {
			fmt.Fprintf(os.Stderr, f+": ", a...)
		}
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
}

func main() {
	logger := NewLogger()

	clargs := ParseCommandLineArguments()

	configFile, err := os.Open(clargs.ConfigurationFilePath)
	logger.DieIfError(err, "failed to open file (%s) for reading", clargs.ConfigurationFilePath)

	_, err = jobber.ReadConfigurationFrom(configFile)
	logger.DieIfError(err, "failed to process configuration in file (%s)", clargs.ConfigurationFilePath)

	_, err = clargs.NewKubeConnector()
	logger.DieIfError(err, "failed to process kubeconfig")
}
