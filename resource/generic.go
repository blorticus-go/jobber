package resource

import (
	"context"
	"fmt"
	"strings"

	"github.com/blorticus-go/jobber/api"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var dpk = metav1.DeletePropagationForeground
var defaultResourceDeletionOptions = metav1.DeleteOptions{
	PropagationPolicy: &dpk,
}

type Generic struct {
	unstructuredApiObject *unstructured.Unstructured
	groupVersionKind      schema.GroupVersionKind
	assumedResource       schema.GroupVersionResource
	client                *api.Client
}

func GuessResourceFromKind(kind string) string {
	return fmt.Sprintf("%ss", strings.ToLower(kind))
}

func NewFromUnstructuredApiObject(u *unstructured.Unstructured, client *api.Client) *Generic {
	gvk := u.GroupVersionKind()
	return &Generic{*unstructuredApiObject: u,
		groupVersionKind: gvk,
		assumedResource: schema.GroupVersionResource{
			Group:    gvk.Group,
			Version:  gvk.Version,
			Resource: GuessResourceFromKind(gvk.Kind),
		},
		client: client,
	}
}

func (r *Generic) Name() string {
	return r.unstructuredApiObject.GetName()
}

func (r *Generic) NamespaceName() string {
	return r.unstructuredApiObject.GetNamespace()
}

func (r *Generic) GroupVersionKind() schema.GroupVersionKind {
	return r.groupVersionKind
}

func (r *Generic) UnstructuredApiObject() *unstructured.Unstructured {
	return r.unstructuredApiObject
}

func (r *Generic) Create() (err error) {
	r.unstructuredApiObject, err = r.client.Dynamic().Resource(r.assumedResource).Namespace(r.NamespaceName()).Create(context.Background(), r.unstructuredApiObject, metav1.CreateOptions{})
	return err
}

func (r *Generic) UpdateStatus() (err error) {
	r.unstructuredApiObject, err = r.client.Dynamic().Resource(r.assumedResource).Namespace(r.NamespaceName()).Get(context.Background(), r.Name(), metav1.GetOptions{})
	return err
}

func (r *Generic) Delete() (err error) {
	return r.client.Dynamic().Resource(r.assumedResource).Namespace(r.NamespaceName()).Delete(context.Background(), r.Name(), defaultResourceDeletionOptions)
}

func (r *Generic) AsAPod() *Pod {
	if SimplifiedTypeStringForUnstructured(r.unstructuredApiObject) != "Pod" {
		return nil
	}

	return &Pod{
		a:              r,
		typedApiObject: new(corev1.Pod),
	}
}

func (r *Generic) AsAJob() *Job {
	if SimplifiedTypeStringForUnstructured(r.unstructuredApiObject) != "Job" {
		return nil
	}

	return &Job{
		a:              r,
		typedApiObject: new(batchv1.Job),
	}
}
