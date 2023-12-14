package jobber

import "github.com/qdm12/reprint"

type PipelineVariables struct {
	Values map[string]any
	Config *TemplateExpansionConfigVariables
}

func NewEmptyPipelineVariables() *PipelineVariables {
	return &PipelineVariables{
		Values: make(map[string]any),
		Config: &TemplateExpansionConfigVariables{
			Namespaces: make(map[string]*TemplateExpansionNamespace),
		},
	}
}

func NewPipelineVariablesWithSeedValues(seedValues map[string]any) *PipelineVariables {
	p := NewEmptyPipelineVariables()

	for key, value := range seedValues {
		p.Values[key] = reprint.This(value)
	}

	return p
}

func (v *PipelineVariables) DeepCopy() *PipelineVariables {
	return reprint.This(v).(*PipelineVariables)
}

func (v *PipelineVariables) AddNamespaceToConfig(namespaceLabel string, namespaceName string) *PipelineVariables {
	v.Config.Namespaces[namespaceLabel] = &TemplateExpansionNamespace{GeneratedName: namespaceName}
	return v
}

func (v *PipelineVariables) MergeValuesToCopy(mergeInValues map[string]any) *PipelineVariables {
	c := v.DeepCopy()

	for key, value := range mergeInValues {
		c.Values[key] = reprint.This(value)
	}

	return c
}
