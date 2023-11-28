package jobber

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

type RunConfigurationNamespacesYaml struct {
	ReferenceName string `yaml:"ReferenceName"`
	BaseName      string `yaml:"BaseName"`
}

type RunConfigurationYaml struct {
	Namespaces       []*RunConfigurationNamespacesYaml `yaml:"Namespaces"`
	DefaultNamespace string                            `yaml:"DefaultNamespace"`
	Pipeline         []string                          `yaml:"Pipeline"`
}

func (c *RunConfigurationYaml) namespacesDoNotHaveNamespaceNamed(name string) bool {
	for _, ns := range c.Namespaces {
		if ns.ReferenceName == name {
			return false
		}
	}

	return true
}

type ConfigurationYaml struct {
	RunConfiguration *RunConfigurationYaml `yaml:"RunConfiguration"`
}

func (c *ConfigurationYaml) validate() error {
	if c.RunConfiguration != nil {
		switch c.RunConfiguration.DefaultNamespace {
		case "":
		case "jobber":
		default:
			switch {
			case c.RunConfiguration.Namespaces == nil:
				return fmt.Errorf("a DefaultNamespace (%s) was provided but there is no matching Namespaces definition", c.RunConfiguration.DefaultNamespace)
			case c.RunConfiguration.namespacesDoNotHaveNamespaceNamed(c.RunConfiguration.DefaultNamespace):
				return fmt.Errorf("the default namespace (%s) is not defined as a Namespace", c.RunConfiguration.DefaultNamespace)
			}
		}
	}

	return nil
}

func ReadConfigurationFrom(r io.Reader) (*ConfigurationYaml, error) {
	c := &ConfigurationYaml{}

	encoder := yaml.NewDecoder(r)
	err := encoder.Decode(c)

	if err := c.validate(); err != nil {
		return nil, err
	}

	return c, err
}
