package wrapped

import (
	"context"
	"fmt"

	"github.com/blorticus-go/jobber/api"
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

func (g *Generic) Name() string {
	return g.unstructuredApiObject.GetName()
}

func (g *Generic) NamespaceName() string {
	return g.unstructuredApiObject.GetNamespace()
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
	Kind:    "namespaces",
}

func NewPodFromGeneric(g *Generic, client *api.Client) (*Pod, error) {
	if g.IsNotA(podGvk) {
		return nil, fmt.Errorf("requested type is (%s) not (v1/Pod)", g.GroupVersionKindAsAString())
	}

	typedApiObject := new(corev1.Pod)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(g.UnstructuredMap(), typedApiObject); err != nil {
		return nil, err
	}

	return &Pod{
		typedApiObject: typedApiObject,
		client:         client,
	}, nil
}

func NewServiceAccountFromGeneric(g *Generic, client *api.Client) (*ServiceAccount, error) {
	if g.IsNotA(serviceAccountGvk) {
		return nil, fmt.Errorf("requested type is (%s) not (v1/ServiceAccount)", g.GroupVersionKindAsAString())
	}

	typedApiObject := new(corev1.ServiceAccount)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(g.UnstructuredMap(), typedApiObject); err != nil {
		return nil, err
	}

	return &ServiceAccount{
		typedApiObject: typedApiObject,
		client:         client,
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

type ServiceAccount struct {
	typedApiObject *corev1.ServiceAccount
	client         *api.Client
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

func NewNamespaceUsingGeneratedName(basename string, client *api.Client) (*Namespace, error) {
	typedApiObject, err := client.Set().CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{GenerateName: basename}}, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return &Namespace{
		typedApiObject: typedApiObject,
		client:         client,
	}, nil
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
