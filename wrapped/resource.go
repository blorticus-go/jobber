package wrapped

import (
	"context"
	"fmt"
	"time"

	"github.com/blorticus-go/jobber/api"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Resource interface {
	Name() string
	NamespaceName() string
	GroupVersionKind() schema.GroupVersionKind
	GroupVersionResource() schema.GroupVersionResource
	Create() error
	Delete() error
	UpdateStatus() error
	UnstructuredApiObject() *unstructured.Unstructured
	UnstructuredMap() map[string]any
}

type Generic struct {
	unstructuredApiObject *unstructured.Unstructured
	client                *api.Client
	groupVersionResource  *schema.GroupVersionResource
	groupVersionKind      *schema.GroupVersionKind
}

func NewResourceFromMap(objectMap map[string]any, client *api.Client) *Generic {
	return &Generic{
		unstructuredApiObject: &unstructured.Unstructured{
			Object: objectMap,
		},
		client: client,
	}
}

func NewResourceForNamespaceFromMap(objectMap map[string]any, inNamespaceName string, client *api.Client) *Generic {
	u := &unstructured.Unstructured{
		Object: objectMap,
	}

	u.SetNamespace(inNamespaceName)

	return &Generic{
		unstructuredApiObject: u,
		client:                client,
	}
}

func (g *Generic) Name() string {
	return g.unstructuredApiObject.GetName()
}

func (g *Generic) NamespaceName() string {
	return g.unstructuredApiObject.GetNamespace()
}

func (g *Generic) SetNamespaceWithoutCommit(namespaceName string) *Generic {
	g.unstructuredApiObject.SetNamespace(namespaceName)
	return g
}

func (g *Generic) GroupVersionKind() schema.GroupVersionKind {
	if g.groupVersionKind == nil {
		gvk := g.unstructuredApiObject.GroupVersionKind()
		g.groupVersionKind = &gvk
	}

	return *g.groupVersionKind
}

func (g *Generic) GroupVersionResource() schema.GroupVersionResource {
	if g.groupVersionKind == nil {
		gvk := g.GroupVersionKind()
		gvr, err := g.client.DetermineResourceFromGroupVersionKind(gvk)
		if err != nil {
			panic(fmt.Sprintf("failed to derivce GroupVersionResource from GroupVersionKind {%s/%s/%s}", gvk.Group, gvk.Version, gvk.Kind))
		}
		g.groupVersionResource = &gvr
	}

	return *g.groupVersionResource
}

func (g *Generic) Create() error {
	apiObject, err := g.client.Dynamic().Resource(g.GroupVersionResource()).Create(context.Background(), g.unstructuredApiObject, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	g.unstructuredApiObject = apiObject
	return nil
}

func (g *Generic) Delete() error {
	return g.client.Dynamic().Resource(g.GroupVersionResource()).Delete(context.Background(), g.Name(), g.client.DefaultResourceDeletionOptions())
}

func (g *Generic) UpdateStatus() error {
	apiObject, err := g.client.Dynamic().Resource(g.GroupVersionResource()).Get(context.Background(), g.Name(), metav1.GetOptions{})
	if err != nil {
		return err
	}

	g.unstructuredApiObject = apiObject
	return nil
}

func (g *Generic) UnstructuredApiObject() *unstructured.Unstructured {
	return g.unstructuredApiObject
}

func (g *Generic) UnstructuredMap() map[string]any {
	return g.unstructuredApiObject.Object
}

func (g *Generic) IsA(gvk schema.GroupVersionKind) bool {
	myGvk := g.GroupVersionKind()
	return myGvk.Group == gvk.Group && myGvk.Version == gvk.Version && myGvk.Kind == gvk.Kind
}

func (g *Generic) IsNotA(gvk schema.GroupVersionKind) bool {
	return !g.IsA(gvk)
}

func (g *Generic) GroupVersionKindAsAString() string {
	gvk := g.GroupVersionKind()
	if gvk.Group == "" {
		return fmt.Sprintf("%s/%s", gvk.Version, gvk.Kind)
	}

	return fmt.Sprintf("%s/%s/%s", gvk.Group, gvk.Version, gvk.Kind)
}

type Pod struct {
	typedApiObject *corev1.Pod
	client         *api.Client
}

var podGvk = schema.GroupVersionKind{
	Group:   "",
	Version: "v1",
	Kind:    "Pod",
}

var serviceAccountGvk = schema.GroupVersionKind{
	Group:   "",
	Version: "v1",
	Kind:    "ServiceAccount",
}

var namespaceGvk = schema.GroupVersionKind{
	Group:   "",
	Version: "v1",
	Kind:    "Namespace",
}

var jobGvk = schema.GroupVersionKind{
	Group:   "batch",
	Version: "v1",
	Kind:    "Job",
}

func NewPodFromGeneric(g *Generic) (*Pod, error) {
	if g.IsNotA(podGvk) {
		return nil, fmt.Errorf("requested type is (%s) not (v1/Pod)", g.GroupVersionKindAsAString())
	}

	typedApiObject := new(corev1.Pod)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(g.UnstructuredMap(), typedApiObject); err != nil {
		return nil, err
	}

	return &Pod{
		typedApiObject: typedApiObject,
		client:         g.client,
	}, nil
}

func (p *Pod) Name() string {
	return p.typedApiObject.Name
}

func (p *Pod) NamespaceName() string {
	return p.typedApiObject.Namespace
}

func (p *Pod) GroupVersionKind() schema.GroupVersionKind {
	return podGvk
}

func (p *Pod) GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}
}

func (p *Pod) Create() error {
	updatedApiObject, err := p.client.Set().CoreV1().Pods(p.typedApiObject.Namespace).Create(context.Background(), p.typedApiObject, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	p.typedApiObject = updatedApiObject
	return nil
}

func (p *Pod) Delete() error {
	return p.client.Set().CoreV1().Pods(p.typedApiObject.Namespace).Delete(context.Background(), p.typedApiObject.Name, p.client.DefaultResourceDeletionOptions())
}

func (p *Pod) UpdateStatus() error {
	updatedApiObject, err := p.client.Set().CoreV1().Pods(p.typedApiObject.Namespace).Get(context.Background(), p.typedApiObject.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	p.typedApiObject = updatedApiObject
	return nil
}

func (p *Pod) UnstructuredApiObject() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: p.UnstructuredMap(),
	}
}

func (p *Pod) UnstructuredMap() map[string]any {
	uMap, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(p.typedApiObject)
	return uMap
}

func (p *Pod) TypedApiObject() *corev1.Pod {
	return p.typedApiObject
}

func (p *Pod) PodIpAsAString() string {
	return p.typedApiObject.Status.PodIP
}

func (p *Pod) WaitForRunningState(maximumAmountOfTimeToWait time.Duration) error {
	timer := NewWaitTimer(maximumAmountOfTimeToWait, time.Second)

	return timer.TestExpectation(
		p,
		func(objectToTest Updatable) (expectationReached bool, errorOccurred error) {
			return p.typedApiObject.Status.Phase == corev1.PodRunning, nil
		},
	)
}

type ServiceAccount struct {
	typedApiObject *corev1.ServiceAccount
	client         *api.Client
}

func NewServiceAccountFromGeneric(g *Generic) (*ServiceAccount, error) {
	if g.IsNotA(serviceAccountGvk) {
		return nil, fmt.Errorf("requested type is (%s) not (v1/ServiceAccount)", g.GroupVersionKindAsAString())
	}

	typedApiObject := new(corev1.ServiceAccount)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(g.UnstructuredMap(), typedApiObject); err != nil {
		return nil, err
	}

	return &ServiceAccount{
		typedApiObject: typedApiObject,
		client:         g.client,
	}, nil
}

func (sa *ServiceAccount) Name() string {
	return sa.typedApiObject.Name
}

func (sa *ServiceAccount) NamespaceName() string {
	return sa.typedApiObject.Namespace
}

func (sa *ServiceAccount) GroupVersionKind() schema.GroupVersionKind {
	return serviceAccountGvk
}

func (sa *ServiceAccount) GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "serviceaccounts",
	}
}

func (sa *ServiceAccount) Create() error {
	updatedTypedApiObject, err := sa.client.Set().CoreV1().ServiceAccounts(sa.typedApiObject.Namespace).Create(context.Background(), sa.typedApiObject, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	sa.typedApiObject = updatedTypedApiObject
	return nil
}

func (sa *ServiceAccount) Delete() error {
	return sa.client.Set().CoreV1().ServiceAccounts(sa.typedApiObject.Namespace).Delete(context.Background(), sa.typedApiObject.Name, sa.client.DefaultResourceDeletionOptions())
}

func (sa *ServiceAccount) UpdateStatus() error {
	updatedTypedApiObject, err := sa.client.Set().CoreV1().ServiceAccounts(sa.typedApiObject.Namespace).Get(context.Background(), sa.typedApiObject.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	sa.typedApiObject = updatedTypedApiObject
	return nil
}

func (sa *ServiceAccount) UnstructuredApiObject() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: sa.UnstructuredMap(),
	}
}

func (sa *ServiceAccount) UnstructuredMap() map[string]any {
	uMap, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(sa.typedApiObject)
	return uMap
}

func (sa *ServiceAccount) TypedApiObject() *corev1.ServiceAccount {
	return sa.typedApiObject
}

type Namespace struct {
	typedApiObject *corev1.Namespace
	client         *api.Client
}

func (n *Namespace) TypedApiObject() *corev1.Namespace {
	return n.typedApiObject
}

func NewNamespaceUsingGeneratedName(basename string, client *api.Client) *Namespace {
	return &Namespace{
		client: client,
		typedApiObject: &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: basename,
			},
		},
	}
}

func (n *Namespace) Name() string {
	return n.typedApiObject.Name
}

func (n *Namespace) NamespaceName() string {
	return n.typedApiObject.Namespace
}

func (n *Namespace) GroupVersionKind() schema.GroupVersionKind {
	return namespaceGvk
}

func (n *Namespace) GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "namespaces",
	}
}

func (n *Namespace) Create() error {
	updatedApiObject, err := n.client.Set().CoreV1().Namespaces().Create(context.Background(), n.typedApiObject, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	n.typedApiObject = updatedApiObject
	return nil
}

func (n *Namespace) Delete() error {
	return n.client.Set().CoreV1().Namespaces().Delete(context.Background(), n.typedApiObject.Name, n.client.DefaultResourceDeletionOptions())
}

func (n *Namespace) UpdateStatus() error {
	updatedApiObject, err := n.client.Set().CoreV1().Namespaces().Get(context.Background(), n.typedApiObject.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	n.typedApiObject = updatedApiObject
	return nil
}

func (n *Namespace) UnstructuredApiObject() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: n.UnstructuredMap(),
	}
}

func (n *Namespace) UnstructuredMap() map[string]any {
	uMap, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(n.typedApiObject)
	return uMap
}

type Job struct {
	typedApiObject *batchv1.Job
	client         *api.Client
}

func (resource *Job) TypedApiObject() *batchv1.Job {
	return resource.typedApiObject
}

func NewJobFromGeneric(g *Generic) (*Job, error) {
	if g.IsNotA(jobGvk) {
		return nil, fmt.Errorf("requested type is (%s) not (batch/v1/Job)", g.GroupVersionKindAsAString())
	}

	typedApiObject := new(batchv1.Job)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(g.UnstructuredMap(), typedApiObject); err != nil {
		return nil, err
	}

	return &Job{
		typedApiObject: typedApiObject,
		client:         g.client,
	}, nil
}

func (resource *Job) Name() string {
	return resource.typedApiObject.Name
}

func (resource *Job) NamespaceName() string {
	return resource.typedApiObject.Namespace
}

func (resource *Job) GroupVersionKind() schema.GroupVersionKind {
	return namespaceGvk
}

func (resource *Job) GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "batch",
		Version:  "v1",
		Resource: "jobs",
	}
}

func (resource *Job) Create() error {
	updatedApiObject, err := resource.client.Set().BatchV1().Jobs(resource.typedApiObject.Namespace).Create(context.Background(), resource.typedApiObject, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	resource.typedApiObject = updatedApiObject
	return nil
}

func (resource *Job) Delete() error {
	return resource.client.Set().BatchV1().Jobs(resource.typedApiObject.Namespace).Delete(context.Background(), resource.typedApiObject.Name, resource.client.DefaultResourceDeletionOptions())
}

func (resource *Job) UpdateStatus() error {
	updatedApiObject, err := resource.client.Set().BatchV1().Jobs(resource.typedApiObject.Namespace).Get(context.Background(), resource.typedApiObject.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	resource.typedApiObject = updatedApiObject
	return nil
}

func (resource *Job) UnstructuredApiObject() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: resource.UnstructuredMap(),
	}
}

func (resource *Job) UnstructuredMap() map[string]any {
	uMap, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(resource.typedApiObject)
	return uMap
}

func (resource *Job) WaitForJobCompletion() error {
	ticker := time.NewTicker(10 * time.Second)

	for {
		if err := resource.UpdateStatus(); err != nil {
			return err
		}

		if resource.typedApiObject.Status.CompletionTime != nil {
			return nil
		}

		if resource.typedApiObject.Status.Failed > 0 {
			return fmt.Errorf("[%d] Pods for the Job failed", resource.typedApiObject.Status.Failed)
		}

		<-ticker.C
	}
}
