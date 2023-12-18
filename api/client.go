package api

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (client *Client) Dynamic() *dynamic.DynamicClient {
	return client.dynamicClient
}

func (client *Client) Set() *kubernetes.Clientset {
	return client.clientSet
}

// func GuessResourceFromKind(kind string) string {
// 	return fmt.Sprintf("%ss", strings.ToLower(kind))
// }

// func (client *Client) CreateNamespaceUsingGeneratedName(generatedBaseName string) (*corev1.Namespace, error) {
// 	apiObject := &corev1.Namespace{
// 		ObjectMeta: metav1.ObjectMeta{
// 			GenerateName: fmt.Sprintf("%s-", generatedBaseName),
// 		},
// 	}

// 	return client.clientSet.CoreV1().Namespaces().Create(context.Background(), apiObject, metav1.CreateOptions{})
// }

// func (client *Client) DeleteNamespace(namespaceApiObject *corev1.Namespace) error {
// 	return client.clientSet.CoreV1().Namespaces().Delete(context.Background(), namespaceApiObject.Name, defaultResourceDeletionOptions)
// }

// func (client *Client) CreateResourceFromUnstructured(instance *unstructured.Unstructured) (*unstructured.Unstructured, error) {
// 	gkv := instance.GroupVersionKind()

// 	gkvResource := schema.GroupVersionResource{
// 		Group:    gkv.Group,
// 		Version:  gkv.Version,
// 		Resource: GuessResourceFromKind(gkv.Kind),
// 	}

// 	return client.dynamicClient.Resource(gkvResource).Namespace(instance.GetNamespace()).Create(context.Background(), instance, metav1.CreateOptions{})
// }

// func (client *Client) UpdateStatusForUnstructured(instance *unstructured.Unstructured) (*unstructured.Unstructured, error) {
// 	gkv := instance.GroupVersionKind()

// 	gkvResource := schema.GroupVersionResource{
// 		Group:    gkv.Group,
// 		Version:  gkv.Version,
// 		Resource: GuessResourceFromKind(gkv.Kind),
// 	}

// 	return client.dynamicClient.Resource(gkvResource).Namespace(instance.GetNamespace()).Get(context.Background(), instance.GetName(), metav1.GetOptions{})
// }

// func (client *Client) DeleteResourceFromUnstructured(instance *unstructured.Unstructured) error {
// 	gkv := instance.GroupVersionKind()

// 	gkvResource := schema.GroupVersionResource{
// 		Group:    gkv.Group,
// 		Version:  gkv.Version,
// 		Resource: GuessResourceFromKind(gkv.Kind),
// 	}

// 	return client.dynamicClient.Resource(gkvResource).Namespace(instance.GetNamespace()).Delete(context.Background(), instance.GetName(), defaultResourceDeletionOptions)
// }