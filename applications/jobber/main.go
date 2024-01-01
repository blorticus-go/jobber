package main

import (
	"github.com/blorticus-go/jobber"
	"github.com/blorticus-go/jobber/api"
)

func main() {
	logger := NewLogger()

	clargs := ParseCommandLineArguments()

	kubeconfigFilePath, err := clargs.ResolveKubeconfigPath()
	logger.DieIfError(err, "cannot resolve kubeconfig file path")

	client, err := api.NewClientUsingKubeconfigFile(kubeconfigFilePath)
	logger.DieIfError(err, "failed to process kubeconfig")

	config, err := jobber.ReadConfigurationYamlFromFile(clargs.ConfigurationFilePath)
	logger.DieIfError(err)

	logger.SetContextFieldWidth(config.CharactersInLongestUnitName(), config.CharactersInLongestCaseName())

	runner := jobber.NewTestRunner(config, client)

	eventChannel := make(chan *jobber.Event)

	go runner.RunTest(eventChannel)

	for {
		event := <-eventChannel
		logger.LogEventMessage(event)

		if event.Type == jobber.TestCompletedSuccessfully || event.Error != nil {
			break
		}
	}
}
