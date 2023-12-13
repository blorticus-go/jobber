package jobber

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type Client struct {
	restConfig    *rest.Config
	clientSet     *kubernetes.Clientset
	dynamicClient *dynamic.DynamicClient
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

	return &Client{
		restConfig:    config,
		clientSet:     clientSet,
		dynamicClient: dynamicClient,
	}, nil
}

func (client *Client) CreateNamespaceUsingGeneratedName(generatedBaseName string) (*corev1.Namespace, error) {
	apiObject := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", generatedBaseName),
		},
	}

	return client.clientSet.CoreV1().Namespaces().Create(context.Background(), apiObject, metav1.CreateOptions{})
}

func (client *Client) MapToUnstructured(inputMap map[string]any) (*unstructured.Unstructured, error) {
	candidate := &unstructured.Unstructured{
		Object: inputMap,
	}

	if gkv := candidate.GroupVersionKind(); gkv.Kind == "" {
		return nil, fmt.Errorf(".kind is not defined")
	} else if gkv.Version == "" {
		return nil, fmt.Errorf(".apiVersion is not defined")
	}

	if candidate.GetName() == "" {
		return nil, fmt.Errorf("metadata.name is not defined")
	}

	return candidate, nil
}

func (client *Client) CreateResourceFromUnstructured(instance *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	gkv := instance.GroupVersionKind()

	gkvResource := schema.GroupVersionResource{
		Group:    gkv.Group,
		Version:  gkv.Version,
		Resource: guessResourceFromKind(gkv.Kind),
	}

	return client.dynamicClient.Resource(gkvResource).Namespace(instance.GetNamespace()).Create(context.Background(), instance, metav1.CreateOptions{})
}

func guessResourceFromKind(kind string) string {
	return fmt.Sprintf("%ss", strings.ToLower(kind))
}
