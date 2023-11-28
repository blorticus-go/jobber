package main

import (
	"flag"
	"fmt"
	"os"
)

type CommandLineArguments struct {
	ConfigurationFilePath string
	KubeconfigPath        string
}

func ParseCommandLineArguments() *CommandLineArguments {
	clargs := &CommandLineArguments{}

	flag.StringVar(&clargs.ConfigurationFilePath, "config", "./config.yaml", "YAML configuration file path")
	flag.StringVar(&clargs.KubeconfigPath, "kubeconfig", "", "kubeconfig file path, if using")
	flag.Parse()

	return clargs
}

func (clargs *CommandLineArguments) NewKubeConnector() (*KubeConnector, error) {
	kubeconfigPath, err := clargs.resolveKubeconfigPath(clargs)
	if err != nil {
		return nil, err
	}

	return NewKubeConnectorFromKubeconfigFile(kubeconfigPath)
}

func (clargs *CommandLineArguments) resolveKubeconfigPath(cliArgs *CommandLineArguments) (string, error) {
	if cliArgs.KubeconfigPath == "" {
		if kubeconfigPathFromEnv := os.Getenv("KUBECONFIG"); kubeconfigPathFromEnv != "" {
			return kubeconfigPathFromEnv, nil
		}
		return "", fmt.Errorf("KUBECONFIG is not set and no kubeconfig path provided on command-line")
	}

	return cliArgs.KubeconfigPath, nil
}
