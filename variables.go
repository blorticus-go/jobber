package jobber

import (
	"context"
	"fmt"

	"github.com/qdm12/reprint"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	client        *Client
}

func NewEmptyPipelineRuntimeValues(client *Client) *PipelineRuntimeValues {
	return &PipelineRuntimeValues{
		createdAssets: make(map[gvkKey]map[resourceName]*GenericK8sResource),
		client:        client,
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

func (values *PipelineRuntimeValues) ServiceAccount(inNamespace string, accountName string) (*TransitiveServiceAccount, error) {
	apiObject, err := values.client.Set().CoreV1().ServiceAccounts(inNamespace).Get(context.Background(), accountName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &TransitiveServiceAccount{
		apiObject: apiObject,
		client:    values.client,
	}, nil
}

type PipelineVariables struct {
	Values  map[string]any
	Config  *TemplateExpansionConfigVariables
	Runtime *PipelineRuntimeValues
}

func NewEmptyPipelineVariables(client *Client) *PipelineVariables {
	return &PipelineVariables{
		Values: make(map[string]any),
		Config: &TemplateExpansionConfigVariables{
			DefaultNamespace: &TemplateExpansionNamespace{},
		},
		Runtime: NewEmptyPipelineRuntimeValues(client),
	}
}

func NewPipelineVariablesWithSeedValues(seedValues map[string]any, client *Client) *PipelineVariables {
	p := NewEmptyPipelineVariables(client)

	for key, value := range seedValues {
		p.Values[key] = reprint.This(value)
	}

	return p
}

func (v *PipelineVariables) DeepCopy() *PipelineVariables {
	return reprint.This(v).(*PipelineVariables)
}

func (v *PipelineVariables) AddDefaultNamespaceToConfig(generatedNamespaceName string) *PipelineVariables {
	v.Config.DefaultNamespace = &TemplateExpansionNamespace{GeneratedName: generatedNamespaceName}
	return v
}

func (v *PipelineVariables) MergeValuesToCopy(mergeInValues map[string]any) *PipelineVariables {
	c := v.DeepCopy()

	for key, value := range mergeInValues {
		c.Values[key] = reprint.This(value)
	}

	return c
}
