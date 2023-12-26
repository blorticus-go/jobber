package wrapped_test

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type MockResource struct {
	R_Name                  string
	R_NamespaceName         string
	R_GroupVersionKind      schema.GroupVersionKind
	R_GroupVersionResource  schema.GroupVersionResource
	R_CreateError           error
	R_DeleteError           error
	R_UpdateStatusError     error
	R_UnstructuredApiObject *unstructured.Unstructured
}

func (r *MockResource) Name() string {
	return r.R_Name
}

func (r *MockResource) NamespaceName() string {
	return r.R_NamespaceName
}

func (r *MockResource) GroupVersionKind() schema.GroupVersionKind {
	return r.R_GroupVersionKind
}

func (r *MockResource) GroupVersionResource() schema.GroupVersionResource {
	return r.R_GroupVersionResource
}

func (r *MockResource) Create() error {
	return r.R_CreateError
}

func (r *MockResource) Delete() error {
	return r.R_DeleteError
}

func (r *MockResource) UpdateStatus() error {
	return r.R_UpdateStatusError
}

func (r *MockResource) UnstructuredApiObject() *unstructured.Unstructured {
	return r.R_UnstructuredApiObject
}

func (r *MockResource) UnstructuredMap() map[string]any {
	return r.R_UnstructuredApiObject.Object
}
