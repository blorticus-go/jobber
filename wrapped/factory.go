package wrapped

import (
	"fmt"

	"github.com/blorticus-go/jobber/api"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type ResourceFactory interface {
	NewResourceFromMap(objectMap map[string]any) Resource
	NewResourceForNamespaceFromMap(objectMap map[string]any, inNamespaceName string) Resource
	NewNamespaceUsingGeneratedName(basename string) Resource
	CoerceResourceToPod(r Resource) (PodResource, error)
	CoerceResourceToJob(r Resource) (JobResource, error)
}

type Factory struct {
	client *api.Client
}

func NewFactory(client *api.Client) *Factory {
	return &Factory{
		client: client,
	}
}

func (factory *Factory) NewResourceFromMap(objectMap map[string]any) Resource {
	return &Generic{
		unstructuredApiObject: &unstructured.Unstructured{
			Object: objectMap,
		},
		client: factory.client,
	}
}

func (factory *Factory) NewResourceForNamespaceFromMap(objectMap map[string]any, inNamespaceName string) Resource {
	u := &unstructured.Unstructured{
		Object: objectMap,
	}

	u.SetNamespace(inNamespaceName)

	return &Generic{
		unstructuredApiObject: u,
		client:                factory.client,
	}
}

func (factory *Factory) NewNamespaceUsingGeneratedName(basename string) Resource {
	return &Namespace{
		client: factory.client,
		typedApiObject: &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: basename,
			},
		},
	}
}

func (factory *Factory) CoerceResourceToPod(r Resource) (PodResource, error) {
	if r.IsNotA(podGvk) {
		return nil, fmt.Errorf("attempt to coerce Resource to Pod but Resource type is (%s)", GroupVersionKindAsAString(r.GroupVersionKind()))
	}

	typedApiObject := new(corev1.Pod)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(r.UnstructuredMap(), typedApiObject); err != nil {
		return nil, fmt.Errorf("attempt to coerce Resource to Pod but unstructured conversion failed: %s", err)
	}

	return &Pod{
		typedApiObject: typedApiObject,
		client:         factory.client,
	}, nil
}

func (factory *Factory) CoerceResourceToJob(r Resource) (JobResource, error) {
	if r.IsNotA(jobGvk) {
		return nil, fmt.Errorf("attempt to coerce Resource to Job but Resource type is (%s)", GroupVersionKindAsAString(r.GroupVersionKind()))
	}

	typedApiObject := new(batchv1.Job)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(r.UnstructuredMap(), typedApiObject); err != nil {
		return nil, fmt.Errorf("attempt to coerce Resource to Pod but unstructured conversion failed: %s", err)
	}

	return &Job{
		typedApiObject: typedApiObject,
		client:         factory.client,
	}, nil
}
