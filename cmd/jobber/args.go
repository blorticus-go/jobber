package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
)

type ConfigVars struct {
	Vars map[string]any
}

func (v *ConfigVars) String() string {
	return ""
}

var configVarMatcher = regexp.MustCompile(`^(\S+)=(.*)$`)

func (v *ConfigVars) Set(s string) error {
	matched := configVarMatcher.FindStringSubmatch(s)
	if len(matched) == 0 {
		return fmt.Errorf("must be varpath=value")
	}

	v.Vars[matched[1]] = matched[2]

	return nil
}

type CommandLineArguments struct {
	ConfigurationFilePath           string
	KubeconfigPath                  string
	OverridenConfigurationVariables map[string]any
}

func ParseCommandLineArguments() *CommandLineArguments {
	clargs := &CommandLineArguments{}

	configVars := &ConfigVars{
		Vars: make(map[string]any),
	}

	flag.StringVar(&clargs.ConfigurationFilePath, "config", "./config.yaml", "YAML configuration file path")
	flag.StringVar(&clargs.KubeconfigPath, "kubeconfig", "", "kubeconfig file path, if using")
	flag.Var(configVars, "set", "add a configuration expansion variable of form varpath=value; may be repeated")
	flag.Parse()

	clargs.OverridenConfigurationVariables = configVars.Vars

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
