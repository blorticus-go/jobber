package resource

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Namespace struct {
	a              *Generic
	typedApiObject *corev1.Pod
}

func (r *Namespace) Name() string {
	return r.a.Name()
}

func (r *Namespace) NamespaceName() string {
	return r.a.NamespaceName()
}

func (r *Namespace) GroupVersionKind() schema.GroupVersionKind {
	return r.a.GroupVersionKind()
}

func (r *Namespace) UnstructuredApiObject() *unstructured.Unstructured {
	return r.a.UnstructuredApiObject()
}

func (r *Namespace) Create() (err error) {
	return r.a.Create()
}

func (r *Namespace) UpdateStatus() (err error) {
	return r.a.UpdateStatus()
}

func (r *Namespace) Delete() (err error) {
	return r.a.Delete()
}

func (r *Namespace) TypedApiObject() *corev1.Pod {
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(r.a.unstructuredApiObject.Object, r.typedApiObject); err != nil {
		return nil
	}

	return r.typedApiObject
}
