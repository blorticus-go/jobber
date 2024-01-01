package pipeline

import (
	"github.com/blorticus-go/jobber/wrapped"
	"github.com/qdm12/reprint"
)

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

func NewVariables() *Variables {
	return &Variables{
		Values: &Values{
			Global: make(map[string]any),
			Unit:   make(map[string]any),
			Case:   make(map[string]any),
		},
		Config: &Config{},
		Runtime: &Runtime{
			Context: &RuntimeContext{},
		},
	}
}

func (v *Variables) SetGlobalValues(globalValues map[string]any) *Variables {
	v.Values.Global = globalValues
	return v
}

func (v *Variables) SetAssetArchiveFilePath(path string) *Variables {
	v.Config.Archive = &ConfigArchiveInformation{FilePath: path}
	return v
}

func (v *Variables) SetDefaultNamespaceName(namespaceName string) *Variables {
	v.Runtime.DefaultNamespace = &DefaultNamespace{Name: namespaceName}
	return v
}

func (v *Variables) DeepCopy() *Variables {
	return reprint.This(v).(*Variables)
}

func (v *Variables) CopyWithAddedTestUnitValues(testUnitName string, values map[string]any) *Variables {
	vCopy := v.DeepCopy()
	vCopy.Values.Unit = reprint.This(values).(map[string]any)
	vCopy.Runtime.Context.CurrentUnit = &RuntimeContextUnit{Name: testUnitName}
	return vCopy
}

func (v *Variables) CopyWithAddedTestCaseValues(testCaseName string, values map[string]any) *Variables {
	vCopy := v.DeepCopy()
	vCopy.Values.Case = reprint.This(values).(map[string]any)
	vCopy.Runtime.Context.CurrentCase = &RuntimeContextCase{Name: testCaseName}
	return vCopy
}
