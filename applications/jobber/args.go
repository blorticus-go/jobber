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

func (clargs *CommandLineArguments) ResolveKubeconfigPath() (string, error) {
	if clargs.KubeconfigPath == "" {
		if kubeconfigPathFromEnv := os.Getenv("KUBECONFIG"); kubeconfigPathFromEnv != "" {
			return kubeconfigPathFromEnv, nil
		}
		return "", fmt.Errorf("KUBECONFIG is not set and no kubeconfig path provided on command-line")
	}

	return clargs.KubeconfigPath, nil
}
