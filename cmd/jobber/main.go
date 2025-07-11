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

	config, err := jobber.ReadConfigurationYamlFromFile(clargs.ConfigurationFilePath)
	logger.DieIfError(err)

	err = config.MergeOverrideValues(clargs.OverridenConfigurationVariables)
	logger.DieIfError(err, "failed to merge override values into configuration: %s", err)

	logger.SetContextFieldWidth(config.CharactersInLongestUnitName(), config.CharactersInLongestCaseName())

	runner := jobber.NewRunner(config, client)

	eventChannel := make(chan *jobber.Event)

	go runner.RunTest(eventChannel)

	for {
		event := <-eventChannel
		logger.LogEventMessage(event)

		if event.Type == jobber.TestingCompletedSuccesfully || event.Error != nil {
			break
		}
	}
}
