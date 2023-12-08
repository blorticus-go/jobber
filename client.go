package jobber

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type Client struct {
	restConfig *rest.Config
	clientSet  *kubernetes.Clientset
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

	return &Client{
		restConfig: config,
		clientSet:  clientSet,
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
