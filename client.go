package jobber

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var dpk = metav1.DeletePropagationForeground
var defaultResourceDeletionOptions = metav1.DeleteOptions{
	PropagationPolicy: &dpk,
}

type Client struct {
	restConfig      *rest.Config
	clientSet       *kubernetes.Clientset
	dynamicClient   *dynamic.DynamicClient
	discoveryClient *discovery.DiscoveryClient
}

func NewClientUsingKubeconfigFile(filePath string) (*Client, error) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: filePath},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}}).ClientConfig()

	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		restConfig:      config,
		clientSet:       clientSet,
		dynamicClient:   dynamicClient,
		discoveryClient: discoveryClient,
	}, nil
}

func (client *Client) Dynamic() *dynamic.DynamicClient {
	return client.dynamicClient
}

func (client *Client) Set() *kubernetes.Clientset {
	return client.clientSet
}

func (client *Client) Discovery() *discovery.DiscoveryClient {
	return client.discoveryClient
}

func (client *Client) CreateNamespaceUsingGeneratedName(generatedBaseName string) (*corev1.Namespace, error) {
	apiObject := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", generatedBaseName),
		},
	}

	return client.clientSet.CoreV1().Namespaces().Create(context.Background(), apiObject, metav1.CreateOptions{})
}

func (client *Client) DeleteNamespace(named string) error {
	return client.clientSet.CoreV1().Namespaces().Delete(context.Background(), named, defaultResourceDeletionOptions)
}

func (client *Client) DefaultResourceDeletionOptions() metav1.DeleteOptions {
	dpk := metav1.DeletePropagationForeground

	return metav1.DeleteOptions{
		PropagationPolicy: &dpk,
	}
}

func (client *Client) DetermineResourceFromGroupVersionKind(gvk schema.GroupVersionKind) (schema.GroupVersionResource, error) {
	var groupVersionString string

	if gvk.Group == "" {
		groupVersionString = gvk.Version
	} else {
		groupVersionString = fmt.Sprintf("%s/%s", gvk.Group, gvk.Version)
	}

	resources, err := client.discoveryClient.ServerResourcesForGroupVersion(groupVersionString)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	for _, resource := range resources.APIResources {
		if resource.Kind == gvk.Kind {
			return schema.GroupVersionResource{
				Group:    gvk.Group,
				Version:  gvk.Version,
				Resource: resource.Name,
			}, nil
		}
	}

	return schema.GroupVersionResource{}, fmt.Errorf("could not find definition for resource %s/%s", groupVersionString, gvk.Kind)
}
