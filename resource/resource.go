package resource

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Type interface {
	Name() string
	NamespaceName() string
	GroupVersionKind() schema.GroupVersionKind
	UnstructuredApiMap() map[string]any
	Create() error
	UpdateStatus() error
	Delete() error
}

type gvkKey string

var gkvToSimplifiedType = map[gvkKey]string{
	gvkKey("\tv1\tPod"):       "Pod",
	gvkKey("batch\tv1\tJob"):  "Job",
	gvkKey("\tv1\tNamespace"): "Namespace",
}

func SimplifiedTypeStringForUnstructured(u *unstructured.Unstructured) string {
	gvk := u.GroupVersionKind()
	lookupKey := fmt.Sprintf("%s\t%s\t%s", gvk.Group, gvk.Version, gvk.Kind)
	return gkvToSimplifiedType[gvkKey(lookupKey)]
}
