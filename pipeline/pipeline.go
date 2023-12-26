package pipeline

import "github.com/blorticus-go/jobber/wrapped"

type Values struct {
	Global map[string]any
	Unit   map[string]any
	Case   map[string]any
}

type ConfigArchiveInformation struct {
	FilePath string
}

type Config struct {
	Archive *ConfigArchiveInformation
}

type RuntimeContextUnit struct {
	Name string
}

type RuntimeContextCase struct {
	Name                         string
	RetrievedAssetsDirectoryPath string
}

type RuntimeContext struct {
	CurrentUnit *RuntimeContextUnit
	CurrentCase *RuntimeContextCase
}

type DefaultNamespace struct {
	Name string
}

func (p *DefaultNamespace) Pod(name string) *wrapped.Pod {
	return nil
}

type Runtime struct {
	DefaultNamespace *DefaultNamespace
	Context          *RuntimeContext
}

func (p *Runtime) Pod(name string, inNamespaceNamed string) *wrapped.Pod {
	return nil
}

func (p *Runtime) ServiceAccount(named string, inNamespaceNamed string) *wrapped.ServiceAccount {
	return nil
}

type Variables struct {
	Values  *Values
	Config  *Config
	Runtime *Runtime
}
