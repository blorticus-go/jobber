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

type PipelineRuntimeNamespace struct {
	Name string
}

type PipelineRuntimeValues struct {
	DefaultNamespace *PipelineRuntimeNamespace
	createdAssets    map[gvkKey]map[resourceName]*GenericK8sResource
	client           *Client
}

func NewEmptyPipelineRuntimeValues(client *Client) *PipelineRuntimeValues {
	return &PipelineRuntimeValues{
		DefaultNamespace: &PipelineRuntimeNamespace{
			Name: "",
		},
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

type PipelineVariablesValues struct {
	Global map[string]any
	Unit   map[string]any
	Case   map[string]any
}

type PipelineVariablesContext struct {
	TestUnitName                         string
	TestCaseName                         string
	TestCaseRetrievedAssetsDirectoryPath string
}

type PipelineVariables struct {
	Values  *PipelineVariablesValues
	Context *PipelineVariablesContext
	Runtime *PipelineRuntimeValues
}

func NewEmptyPipelineVariables(client *Client) *PipelineVariables {
	return &PipelineVariables{
		Values: &PipelineVariablesValues{
			Global: make(map[string]any),
			Unit:   make(map[string]any),
			Case:   make(map[string]any),
		},
		Context: &PipelineVariablesContext{
			TestUnitName:                         "",
			TestCaseName:                         "",
			TestCaseRetrievedAssetsDirectoryPath: "",
		},
		Runtime: NewEmptyPipelineRuntimeValues(client),
	}
}

func (v *PipelineVariables) DeepCopy() *PipelineVariables {
	return reprint.This(v).(*PipelineVariables)
}

func (v *PipelineVariables) WithGlobalValues(globalValues map[string]any) *PipelineVariables {
	v.Values.Global = globalValues
	return v
}

func (v *PipelineVariables) SetDefaultNamespaceNameTo(generatedNamespaceName string) *PipelineVariables {
	v.Runtime.DefaultNamespace.Name = generatedNamespaceName
	return v
}

func (v *PipelineVariables) AndUsingDefaultNamespaceNamed(generatedNamespaceName string) *PipelineVariables {
	return v.SetDefaultNamespaceNameTo(generatedNamespaceName)
}

func (v *PipelineVariables) RescopedToUnitNamed(testUnitName string) *PipelineVariables {
	vCopy := v.DeepCopy()
	vCopy.Values.Unit = map[string]any{}
	vCopy.Values.Case = map[string]any{}
	vCopy.Context = &PipelineVariablesContext{
		TestUnitName: testUnitName,
	}
	return vCopy
}

func (v *PipelineVariables) WithUnitValues(unitValues map[string]any) *PipelineVariables {
	v.Values.Unit = unitValues
	return v
}

func (v *PipelineVariables) RescopedToCaseNamed(testCaseName string) *PipelineVariables {
	vCopy := v.DeepCopy()
	vCopy.Values.Case = map[string]any{}
	vCopy.Context = &PipelineVariablesContext{
		TestUnitName: v.Context.TestUnitName,
		TestCaseName: testCaseName,
	}
	return vCopy
}

func (v *PipelineVariables) WithCaseValues(caseValues map[string]any) *PipelineVariables {
	v.Values.Case = caseValues
	return v
}

func (v *PipelineVariables) SetTestCaseRetrievedAssetsDirectoryPath(path string) *PipelineVariables {
	v.Context.TestCaseRetrievedAssetsDirectoryPath = path
	return v
}

func (v *PipelineVariables) AndTestCaseRetrievedAssetsDirectoryAt(path string) *PipelineVariables {
	return v.SetTestCaseRetrievedAssetsDirectoryPath(path)
}
