package jobber

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type K8sResource interface {
	GetObjectKind() schema.ObjectKind
	GetName() string
	GetNamespace() string
	UpdateStatus() error
}

type K8sUnstructuredResource struct {
	apiObject *unstructured.Unstructured
	client    *Client
}

func NewK8sUnstructuredResource(unstructuredObject *unstructured.Unstructured, client *Client) *K8sUnstructuredResource {
	return &K8sUnstructuredResource{
		apiObject: unstructuredObject,
		client:    client,
	}
}

func NewK8sUnstructuredResourceFromMap(m map[string]any, client *Client) (*K8sUnstructuredResource, error) {
	u, err := mapToUnstructured(m)
	if err != nil {
		return nil, err
	}

	return NewK8sUnstructuredResource(u, client), nil
}

func (resource *K8sUnstructuredResource) GetObjectKind() schema.ObjectKind {
	return resource.apiObject.GetObjectKind()
}

func (resource *K8sUnstructuredResource) GetNamespace() string {
	return resource.apiObject.GetNamespace()
}

func (resource *K8sUnstructuredResource) SetNamespace(namespaceName string) {
	resource.apiObject.SetNamespace(namespaceName)
}

func (resource *K8sUnstructuredResource) GetName() string {
	return resource.apiObject.GetName()
}

func (resource *K8sUnstructuredResource) Create() (err error) {
	resource.apiObject, err = resource.client.CreateResourceFromUnstructured(resource.apiObject)
	return err
}

func (resource *K8sUnstructuredResource) UpdateStatus() (err error) {
	resource.apiObject, err = resource.client.UpdateStatusForUnstructured(resource.apiObject)
	return err
}

func (resource *K8sUnstructuredResource) Delete() error {
	return resource.client.DeleteResourceFromUnstructured(resource.apiObject)
}

func GuessResourceFromKind(kind string) string {
	return fmt.Sprintf("%ss", strings.ToLower(kind))
}

func (resource *K8sUnstructuredResource) ApiObject() *unstructured.Unstructured {
	return resource.apiObject
}

func (resource *K8sUnstructuredResource) UnstructuredMap() map[string]any {
	return resource.apiObject.Object
}

func (resource *K8sUnstructuredResource) Information() *K8sResourceInformation {
	gkv := resource.apiObject.GetObjectKind().GroupVersionKind()

	return &K8sResourceInformation{
		Kind:          gkv.Kind,
		Name:          resource.GetName(),
		NamespaceName: resource.GetNamespace(),
	}
}

var gkvToSimplifiedType = map[gvkKey]string{
	gvkKey("\tv1\tPod"):       "Pod",
	gvkKey("batch\tv1\tJob"):  "Job",
	gvkKey("\tv1\tNamespace"): "Namespace",
}

func (resource *K8sUnstructuredResource) SimplifiedTypeString() string {
	return gkvToSimplifiedType[gvkKeyFromGroupVersionKind(resource.apiObject.GroupVersionKind())]
}

func mapToUnstructured(inputMap map[string]any) (*unstructured.Unstructured, error) {
	candidate := &unstructured.Unstructured{
		Object: inputMap,
	}

	if gkv := candidate.GroupVersionKind(); gkv.Kind == "" {
		return nil, fmt.Errorf(".kind is not defined")
	} else if gkv.Version == "" {
		return nil, fmt.Errorf(".apiVersion is not defined")
	}

	if candidate.GetName() == "" {
		return nil, fmt.Errorf("metadata.name is not defined")
	}

	return candidate, nil
}
