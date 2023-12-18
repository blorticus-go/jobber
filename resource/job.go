package resource

import (
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Job struct {
	a              *Generic
	typedApiObject *batchv1.Job
}

func (r *Job) Name() string {
	return r.a.Name()
}

func (r *Job) NamespaceName() string {
	return r.a.NamespaceName()
}

func (r *Job) GroupVersionKind() schema.GroupVersionKind {
	return r.a.GroupVersionKind()
}

func (r *Job) UnstructuredApiObject() *unstructured.Unstructured {
	return r.a.UnstructuredApiObject()
}

func (r *Job) Create() (err error) {
	return r.a.Create()
}

func (r *Job) UpdateStatus() (err error) {
	return r.a.UpdateStatus()
}

func (r *Job) Delete() (err error) {
	return r.a.Delete()
}

func (r *Job) TypedApiObject() *batchv1.Job {
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(r.a.unstructuredApiObject.Object, r.typedApiObject); err != nil {
		return nil
	}

	return r.typedApiObject
}
