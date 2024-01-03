package jobber

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"gopkg.in/yaml.v3"
)

type ConfigurationDefaultNamespace struct {
	Basename string `yaml:"Basename"`
}

type TestCase struct {
	Name   string         `yaml:"Name"`
	Values map[string]any `yaml:"Values"`
}

type TestUnit struct {
	Name   string         `yaml:"Name"`
	Values map[string]any `yaml:"Values"`
}

type ConfigurationAssetArchive struct {
	FilePath string `yaml:"FilePath"`
}

type ConfigurationPipeline struct {
	ActionDefinitionsRootDirectory string   `yaml:"ActionDefinitionsRootDirectory"`
	ActionsInOrder                 []string `yaml:"ActionsInOrder"`
}

type ConfigurationTest struct {
	AssetArchive     *ConfigurationAssetArchive     `yaml:"AssetArchive"`
	DefaultNamespace *ConfigurationDefaultNamespace `yaml:"DefaultNamespace"`
	GlobalValues     map[string]any                 `yaml:"GlobalValues"`
	Pipeline         *ConfigurationPipeline         `yaml:"Pipeline"`
	Cases            []*TestCase                    `yaml:"Cases"`
	Units            []*TestUnit                    `yaml:"Units"`
}

type Configuration struct {
	Test *ConfigurationTest `yaml:"Test"`
}

func (c *Configuration) validate() error {
	if c.Test == nil {
		return fmt.Errorf(".Test must exist")
	}

	if c.Test.Cases == nil || len(c.Test.Cases) == 0 {
		return fmt.Errorf(".Test.Cases must exist and cannot be an empty list")
	}

	if c.Test.Units == nil || len(c.Test.Units) == 0 {
		return fmt.Errorf(".Test.Units must exist and cannot be an empty list")
	}

	if c.Test.AssetArchive == nil {
		return fmt.Errorf(".Test.AssetArchive must be defined")
	}

	if c.Test.AssetArchive.FilePath == "" {
		return fmt.Errorf(".Test.AssetArchive.FilePath must exist and cannot be the empty string")
	}

	if c.Test.DefaultNamespace == nil {
		return fmt.Errorf(".Test.DefaultNamespace must be defined")
	}

	if c.Test.DefaultNamespace.Basename == "" {
		return fmt.Errorf(".Test.DefaultNamespace.Basename must be defined and cannot be the empty string")
	}

	if c.Test.Pipeline == nil {
		return fmt.Errorf(".Test.Pipeline must be defined and cannot be empty")
	}

	if c.Test.Pipeline.ActionDefinitionsRootDirectory == "" {
		return fmt.Errorf(".Test.Pipeline.ActionDefinitionRootDirectory must be defined and cannot be empty")
	}

	if len(c.Test.Pipeline.ActionsInOrder) == 0 {
		return fmt.Errorf(".Test.Pipeline.ActionsInOrder must have at least one entry")
	}

	for pipelineEntryIndex, value := range c.Test.Pipeline.ActionsInOrder {
		s := strings.Split(value, "/")
		if len(s) != 2 {
			return fmt.Errorf(".Test.Pipeline.ActionsInOrder[%d] must be of format <type>/<target>", pipelineEntryIndex)
		}
		switch s[0] {
		case "resources":
		case "values-transforms":
		case "executables":
		default:
			return fmt.Errorf(".Test.Pipeline.ActionsInOrder[%d] type indicator [%s] is not understood", pipelineEntryIndex, s[0])
		}
	}

	return nil
}

func (c *Configuration) expandDefaults() {
	if c.Test.GlobalValues == nil {
		c.Test.GlobalValues = make(map[string]any)
	}

	for _, testCase := range c.Test.Cases {
		if testCase.Values == nil {
			testCase.Values = make(map[string]any)
		}
	}

	for _, testUnit := range c.Test.Units {
		if testUnit.Values == nil {
			testUnit.Values = make(map[string]any)
		}
	}
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

	c.expandDefaults()

	return c, err
}

func ReadConfigurationYamlFromFile(filePath string, configExpansionVars map[string]string) (*Configuration, error) {
	tmpl, err := template.New(filepath.Base(filePath)).Funcs(sprig.FuncMap()).ParseFiles(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config at (%s): %s", filePath, err)
	}

	configTemplateBuffer := new(bytes.Buffer)

	if err := tmpl.Execute(configTemplateBuffer, configExpansionVars); err != nil {
		return nil, fmt.Errorf("failed to expand template file (%s): %s", filePath, err)
	}

	return ReadConfigurationYamlFromReader(configTemplateBuffer)
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
