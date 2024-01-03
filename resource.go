package jobber

import (
	"context"
	"fmt"
	"strings"
	"time"

	authenticationv1 "k8s.io/api/authentication/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type GenericK8sResource struct {
	Group                 string
	Version               string
	Kind                  string
	Name                  string
	groupVersionResource  schema.GroupVersionResource
	unstructuredApiObject *unstructured.Unstructured
	client                *Client
}

func GuessResourceFromKind(kind string) string {
	return fmt.Sprintf("%ss", strings.ToLower(kind))
}

func NewGenericK8sResourceFromUnstructured(u *unstructured.Unstructured, client *Client) (*GenericK8sResource, error) {
	gvk := u.GroupVersionKind()

	gvr, err := client.DetermineResourceFromGroupVersionKind(gvk)
	if err != nil {
		return nil, err
	}

	return &GenericK8sResource{
		Group:                 gvk.Group,
		Version:               gvk.Version,
		Kind:                  gvk.Kind,
		Name:                  u.GetName(),
		groupVersionResource:  gvr,
		unstructuredApiObject: u,
		client:                client,
	}, nil
}

func NewGenericK8sResourceFromUnstructuredMap(inputMap map[string]any, client *Client) (*GenericK8sResource, error) {
	candidate := &unstructured.Unstructured{
		Object: inputMap,
	}

	if gkv := candidate.GroupVersionKind(); gkv.Kind == "" {
		return nil, fmt.Errorf(".kind is not defined")
	} else if gkv.Version == "" {
		return nil, fmt.Errorf(".apiVersion is not defined")
	}

	if candidate.GetName() == "" && candidate.GetGenerateName() == "" {
		return nil, fmt.Errorf("neither metadata.name nor metadata.generateName is defined")
	}

	return NewGenericK8sResourceFromUnstructured(candidate, client)
}

func (resource *GenericK8sResource) NamespaceName() string {
	return resource.unstructuredApiObject.GetNamespace()
}

func (resource *GenericK8sResource) SetNamespace(namespaceName string) {
	resource.unstructuredApiObject.SetNamespace(namespaceName)
}

func (resource *GenericK8sResource) Create() (err error) {
	updatedResource, err := resource.client.Dynamic().
		Resource(resource.groupVersionResource).
		Namespace(resource.NamespaceName()).
		Create(
			context.Background(),
			resource.unstructuredApiObject,
			metav1.CreateOptions{},
		)

	if err != nil {
		return err
	}

	resource.unstructuredApiObject = updatedResource
	return nil
}

func (resource *GenericK8sResource) UpdateStatus() (err error) {
	updatedResource, err := resource.client.Dynamic().
		Resource(resource.groupVersionResource).
		Namespace(resource.NamespaceName()).
		Get(
			context.Background(),
			resource.Name,
			metav1.GetOptions{},
		)

	if err != nil {
		return err
	}

	resource.unstructuredApiObject = updatedResource
	return nil
}

func (resource *GenericK8sResource) Delete() error {
	return resource.client.Dynamic().
		Resource(resource.groupVersionResource).
		Namespace(resource.NamespaceName()).
		Delete(
			context.Background(),
			resource.Name,
			resource.client.DefaultResourceDeletionOptions(),
		)
}

func (resource *GenericK8sResource) ApiObject() *unstructured.Unstructured {
	return resource.unstructuredApiObject
}

func (resource *GenericK8sResource) UnstructuredMap() map[string]any {
	return resource.unstructuredApiObject.Object
}

func (resource *GenericK8sResource) Information() *K8sResourceInformation {
	return &K8sResourceInformation{
		Kind:          resource.Kind,
		Name:          resource.Name,
		NamespaceName: resource.NamespaceName(),
	}
}

func (resource *GenericK8sResource) GvkString() string {
	if resource.Group == "" {
		return fmt.Sprintf("%s/%s", resource.Version, resource.Kind)
	}

	return fmt.Sprintf("%s/%s/%s", resource.Group, resource.Version, resource.Kind)
}

type TransitivePod struct {
	genericResource *GenericK8sResource
	client          *Client
}

type TransitiveJob struct {
	genericResource *GenericK8sResource
	client          *Client
}

type TransitiveServiceAccount struct {
	apiObject *corev1.ServiceAccount
	client    *Client
}

func (resource *GenericK8sResource) AsAPod() *TransitivePod {
	return &TransitivePod{
		genericResource: resource,
		client:          resource.client,
	}
}

func (resource *GenericK8sResource) AsAJob() *TransitiveJob {
	return &TransitiveJob{
		genericResource: resource,
		client:          resource.client,
	}
}

func (pod *TransitivePod) UpdateStatus() (err error) {
	return pod.genericResource.UpdateStatus()
}

func (pod *TransitivePod) typedApiObject() (*corev1.Pod, error) {
	typed := new(corev1.Pod)
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(pod.genericResource.ApiObject().Object, typed)
	return typed, err
}

func (pod *TransitivePod) WaitForRunningState(lengthOfTimeToWait time.Duration) error {
	timer := NewWaitTimer(lengthOfTimeToWait, time.Second)

	return timer.TestExpectation(
		pod,
		func(objectToTest Updatable) (expectationReached bool, errorOccurred error) {
			podApiObject, err := pod.typedApiObject()
			if err != nil {
				return false, err
			}
			return podApiObject.Status.Phase == corev1.PodRunning, nil
		},
	)
}

func (pod *TransitivePod) IpString() (string, error) {
	podApiObject, err := pod.typedApiObject()
	if err != nil {
		return "", fmt.Errorf("cannot convert generic API object to Pod API object: %s", err)
	}

	return podApiObject.Status.PodIP, nil
}

func (job *TransitiveJob) typedApiObject() (*batchv1.Job, error) {
	typed := new(batchv1.Job)
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(job.genericResource.ApiObject().Object, typed)
	return typed, err

}

func (job *TransitiveJob) updateJobApiObjectStatus(apiObject *batchv1.Job) (*batchv1.Job, error) {
	return job.client.clientSet.BatchV1().Jobs(apiObject.Namespace).Get(context.Background(), apiObject.Name, metav1.GetOptions{})
}

func (job *TransitiveJob) WaitForCompletion() error {
	jobApiObject, err := job.typedApiObject()
	if err != nil {
		return fmt.Errorf("cannot convert generic API object to Job API object: %s", err)
	}

	ticker := time.NewTicker(10 * time.Second)

	for {
		<-ticker.C

		jobApiObject, err := job.updateJobApiObjectStatus(jobApiObject)
		if err != nil {
			return fmt.Errorf("failed to update Job status: %s", err)
		}

		if jobApiObject.Status.CompletionTime != nil {
			return nil
		}

		if jobApiObject.Status.Failed > 0 {
			return fmt.Errorf("[%d] Pods for the Job failed", jobApiObject.Status.Failed)
		}
	}
}

func (sa *TransitiveServiceAccount) GenerateBoundBearerTokenString() (string, error) {
	tokenRequest, err := sa.client.Set().CoreV1().ServiceAccounts(sa.apiObject.Namespace).CreateToken(context.Background(), sa.apiObject.Name, &authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			Audiences: []string{"api", "https://kubernetes.default.svc"},
		},
	}, metav1.CreateOptions{})

	if err != nil {
		return "", err
	}

	return tokenRequest.Status.Token, nil
}
