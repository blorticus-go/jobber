package jobber

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type ConfigurationNamespace struct {
	Basename string `yaml:"Basename"`
}

type ConfigurationDefinition struct {
	DefaultValues         map[string]any                     `yaml:"DefaultValues"`
	Namespaces            map[string]*ConfigurationNamespace `yaml:"Namespaces"`
	PipelineRootDirectory string                             `yaml:"PipelineRootDirectory"`
	Pipeline              []string                           `yaml:"Pipeline"`
}

type TestCase struct {
	Name   string         `yaml:"Name"`
	Values map[string]any `yaml:"Values"`
}

type TestUnit struct {
	Name   string         `yaml:"Name"`
	Values map[string]any `yaml:"Value"`
}

type ConfigurationTest struct {
	Definition *ConfigurationDefinition `yaml:"Definition"`
	Cases      []*TestCase              `yaml:"Cases"`
	Units      []*TestUnit              `yaml:"Units"`
}

type Configuration struct {
	Test *ConfigurationTest `yaml:"Test"`
}

func (c *Configuration) validate() error {
	if c.Test == nil {
		return fmt.Errorf(".Test must exist")
	}

	if c.Test.Cases == nil {
		return fmt.Errorf(".Test.Cases must exist and cannot be an empty list")
	}

	if c.Test.Units == nil {
		return fmt.Errorf(".Test.Units must exist and cannot be an empty list")
	}

	if c.Test.Definition == nil {
		return fmt.Errorf(".Test.Definition must exist")
	}

	if c.Test.Definition.Namespaces == nil {
		return fmt.Errorf(".Test.Definition.Namepsaces must define at least the 'Default' namespace")
	}

	if _, keyIsInMap := c.Test.Definition.Namespaces["Default"]; !keyIsInMap {
		return fmt.Errorf(".Test.Definition.Namepsaces must define at least the 'Default' namespace")
	}

	if c.Test.Definition.Pipeline == nil {
		return fmt.Errorf(".Test.Definition.Pipeline must exist and must not be an empty list")
	}

	if c.Test.Definition.PipelineRootDirectory == "" {
		return fmt.Errorf(".Test.Definition.PipelineRootDirectory must exist and my not be empty")
	}

	for pipelineEntryIndex, value := range c.Test.Definition.Pipeline {
		s := strings.Split(value, "/")
		if len(s) != 2 {
			return fmt.Errorf(".Test.Definition.Pipeline.[%d] must be of format <type>/<target>", pipelineEntryIndex)
		}
		switch s[0] {
		case "resources":
		case "values-transforms":
		case "executables":
		default:
			return fmt.Errorf(".Test.Definition.Pipeline.[%d] type indicator [%s] is not understood", pipelineEntryIndex, s[0])
		}
	}

	return nil
}

func ReadConfigurationYamlFromReader(r io.Reader) (*Configuration, error) {
	c := &Configuration{}

	encoder := yaml.NewDecoder(r)
	err := encoder.Decode(c)

	if err != nil {
		return nil, err
	}

	if err := c.validate(); err != nil {
		return nil, err
	}

	return c, err
}

func ReadConfigurationYamlFromFile(filePath string) (*Configuration, error) {
	yamlFile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file (%s): %s", filePath, err.Error())
	}
	defer yamlFile.Close()

	return ReadConfigurationYamlFromReader(yamlFile)
}

func (c *Configuration) CharactersInLongestCaseName() uint {
	longest := 0
	for _, c := range c.Test.Cases {
		if len(c.Name) > longest {
			longest = len(c.Name)
		}
	}

	return uint(longest)
}

func (c *Configuration) CharactersInLongestUnitName() uint {
	longest := 0
	for _, c := range c.Test.Units {
		if len(c.Name) > longest {
			longest = len(c.Name)
		}
	}

	return uint(longest)
}
