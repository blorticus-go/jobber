package jobber

import (
	"fmt"

	"github.com/qdm12/reprint"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type gvkKey string
type resourceName string

func gvkKeyFromGroupVersionKind(gvk schema.GroupVersionKind) gvkKey {
	return gvkKey(fmt.Sprintf("%s\t%s\t%s", gvk.Group, gvk.Version, gvk.Kind))
}

func gvkKeyFromGVKStrings(group, version, kind string) gvkKey {
	return gvkKey(fmt.Sprintf("%s\t%s\t%s", group, version, kind))
}

type PipelineRuntimeValues struct {
	createdAssets map[gvkKey]map[resourceName]*GenericK8sResource
}

func NewEmptyPipelineRuntimeValues() *PipelineRuntimeValues {
	return &PipelineRuntimeValues{
		createdAssets: make(map[gvkKey]map[resourceName]*GenericK8sResource),
	}
}

func (values *PipelineRuntimeValues) Add(resource *GenericK8sResource) *PipelineRuntimeValues {
	key := gvkKeyFromGroupVersionKind(resource.ApiObject().GroupVersionKind())

	if values.createdAssets[key] == nil {
		values.createdAssets[key] = make(map[resourceName]*GenericK8sResource)
	}

	values.createdAssets[key][resourceName(resource.Name)] = resource

	return values
}

func (values *PipelineRuntimeValues) CreatedAsset(group string, version string, kind string, name string) *GenericK8sResource {
	return values.createdAssets[gvkKeyFromGVKStrings(group, version, kind)][resourceName(name)]
}

func (values *PipelineRuntimeValues) CreatedPod(podName string) (*TransitivePod, error) {
	if resource := values.CreatedAsset("", "v1", "Pod", podName); resource == nil {
		return nil, fmt.Errorf("no created pod named (%s)", podName)
	} else {
		return resource.AsAPod(), nil
	}
}

type PipelineVariables struct {
	Values  map[string]any
	Config  *TemplateExpansionConfigVariables
	Runtime *PipelineRuntimeValues
}

func NewEmptyPipelineVariables() *PipelineVariables {
	return &PipelineVariables{
		Values: make(map[string]any),
		Config: &TemplateExpansionConfigVariables{
			Namespaces: make(map[string]*TemplateExpansionNamespace),
		},
		Runtime: NewEmptyPipelineRuntimeValues(),
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
