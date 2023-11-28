package jobber

type RunConfigurationNamespacesYaml struct {
	ReferenceName string `yaml:"ReferenceName"`
	BaseName      string `yaml:"BaseName"`
}

type RunConfigurationYaml struct {
	Namespaces       RunConfigurationNamespacesYaml `yaml:"Namespaces"`
	DefaultNamespace string                         `yaml:"DefaultNamespace"`
	Pipeline         []string                       `yaml:"Pipeline"`
}

type ConfigurationYaml struct {
	RunConfiguration RunConfigurationYaml `yaml:"RunConfiguration"`
}
