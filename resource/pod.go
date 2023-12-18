package resource

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Pod struct {
	a              *Generic
	typedApiObject *corev1.Pod
}

func (r *Pod) Name() string {
	return r.a.Name()
}

func (r *Pod) NamespaceName() string {
	return r.a.NamespaceName()
}

func (r *Pod) GroupVersionKind() schema.GroupVersionKind {
	return r.a.GroupVersionKind()
}

func (r *Pod) UnstructuredApiObject() *unstructured.Unstructured {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(r.typedApiObject)
	if err != nil {
		return nil
	}

	return &unstructured.Unstructured{
		Object: u,
	}
}

func (r *Pod) Create() (err error) {
	r.typedApiObject, err = r.a.client.Set().CoreV1().Pods(r.typedApiObject.Namespace).Create(context.Background(), r.typedApiObject, metav1.CreateOptions{})
	return err
}

func (r *Pod) UpdateStatus() (err error) {
	r.typedApiObject, err = r.a.client.Set().CoreV1().Pods(r.typedApiObject.Namespace).Get(context.Background(), r.typedApiObject.Name, metav1.GetOptions{})
	return err
}

func (r *Pod) Delete() (err error) {
	return r.a.client.Set().CoreV1().Pods(r.typedApiObject.Namespace).Delete(context.Background(), r.typedApiObject.Name, defaultResourceDeletionOptions)
}

func (r *Pod) TypedApiObject() *corev1.Pod {
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(r.a.unstructuredApiObject.Object, r.typedApiObject); err != nil {
		return nil
	}

	return r.typedApiObject
}

// func (r *Pod) WaitForRunningState(amountOfTimeToWait time.Duration) error {
// 	t := &jobber.WaitTimer{
// 		MaximumTimeToWait: amountOfTimeToWait,
// 		ProbeInterval:     time.Second,
// 	}

// 	return t.TestExpectation(
// 		r,
// 		func(kr K8sResource) (expectationReached bool, errorOccurred error) {
// 			p := pod.ApiObject().Object["status"].(map[string]any)["phase"]
// 			return (kr != nil && p == "Running"), nil
// 		})

// }
