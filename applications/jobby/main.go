package main

import (
	"github.com/blorticus-go/jobber"
)

func main() {
	logger := NewLogger()

	clargs := ParseCommandLineArguments()

	kubeconfigFilePath, err := clargs.ResolveKubeconfigPath()
	logger.DieIfError(err, "cannot resolve kubeconfig file path")

	client, err := jobber.NewClientUsingKubeconfigFile(kubeconfigFilePath)
	logger.DieIfError(err, "failed to process kubeconfig")

	runner := jobber.NewRunner(client)

	err = runner.ReadConfigurationFromFile(clargs.ConfigurationFilePath)
	logger.DieIfError(err)
}
