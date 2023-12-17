package jobber

import (
	"fmt"

	"github.com/qdm12/reprint"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	createdAssets map[gvkKey]map[resourceName]*unstructured.Unstructured
}

func NewEmptyPipelineRuntimeValues() *PipelineRuntimeValues {
	return &PipelineRuntimeValues{
		createdAssets: make(map[gvkKey]map[resourceName]*unstructured.Unstructured),
	}
}

func (values *PipelineRuntimeValues) Add(u *unstructured.Unstructured) *PipelineRuntimeValues {
	key := gvkKeyFromGroupVersionKind(u.GroupVersionKind())

	if values.createdAssets[key] == nil {
		values.createdAssets[key] = make(map[resourceName]*unstructured.Unstructured)
	}

	values.createdAssets[key][resourceName(u.GetName())] = u

	return values
}

func (values *PipelineRuntimeValues) CreatedAsset(group string, version string, kind string, name string) *unstructured.Unstructured {
	return values.createdAssets[gvkKeyFromGVKStrings(group, version, kind)][resourceName(name)]
}

// func (values *PipelineRuntimeValues) CreatedPod(podName string) (*corev1.Pod, error) {
// 	if u := values.CreatedAsset("", "v1", "Pod", podName); u == nil {
// 		return nil, fmt.Errorf("no created pod named (%s)", podName)
// 	} else {
// 		return UnstructuredToPodType(u)
// 	}
// }

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
