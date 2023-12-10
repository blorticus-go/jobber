package jobber

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TemplateExpansionNamespace struct {
	GeneratedName string
}

type TemplateExpansionConfigVariables struct {
	Namespaces map[string]*TemplateExpansionNamespace
}

type TemplateExpansionVariables struct {
	Values map[string]any
	Config *TemplateExpansionConfigVariables
}

type RunningContext struct {
	testUnit *TestUnit
	testCase *TestCase
}

type Runner struct {
	client               *Client
	config               *Configuration
	TemplateExpansions   *TemplateExpansionVariables
	defaultNamespaceName string
	resourceTracker      *CreatedResourceTracker
}

func NewRunner(config *Configuration, client *Client) *Runner {
	return &Runner{
		client:          client,
		config:          config,
		resourceTracker: NewCreatedResourceTracker(),
	}
}

func (runner *Runner) createDefaultNamespace() (*corev1.Namespace, error) {
	defaultNamespaceApiObject, err := runner.client.CreateNamespaceUsingGeneratedName(runner.config.Test.Definition.Namespaces["Default"].Basename)
	if err != nil {
		return nil, err
	}

	runner.resourceTracker.AddCreatedResource(&K8sResource{
		information: &K8sResourceInformation{
			Kind:          "namespace",
			Name:          defaultNamespaceApiObject.Name,
			NamespaceName: "",
		},
		deletionMethod: func(object any) error {
			return runner.client.clientSet.CoreV1().Namespaces().Delete(context.Background(), defaultNamespaceApiObject.Name, metav1.DeleteOptions{})
		},
	})

	return defaultNamespaceApiObject, nil
}

func (runner *Runner) RunTest(eventChannel chan<- *Event) {
	eventHandler := &eventHandler{eventChannel}

	for _, testUnit := range runner.config.Test.Units {
		eventHandler.sayThatUnitStarted(testUnit)

		for _, testCase := range runner.config.Test.Cases {
			eventHandler.sayThatCaseStarted(testUnit, testCase)

			nsObject, err := runner.createDefaultNamespace()
			eventHandler.explainAttemptToCreateDefaultNamespace(runner.config.Test.Definition.Namespaces["Default"].Basename, nsObject, EventContextFor(testUnit, testCase), err)
			if err != nil {
				return
			}

			for _, attemptDetails := range runner.resourceTracker.AttemptToDeleteAllAsYetUndeletedResources() {
				if attemptDetails.Error != nil {
					eventHandler.sayThatResourceDeletionFailed(attemptDetails.Resource.information, attemptDetails.Error, testUnit, testCase)
					return
				}

				eventHandler.sayThatResourceDeletionSucceeded(attemptDetails.Resource.information, testUnit, testCase)
			}

			eventHandler.sayThatCaseCompletedSuccessfully(testUnit, testCase)
		}

		eventHandler.sayThatUnitCompletedSuccessfully(testUnit)
	}

	eventHandler.sayThatTestingCompletedSuccessfully()
}

func (runner *Runner) CreateResourceFromYaml(yamlAsString string) error {
	// client, err := dynamic.NewForConfig(runner.client.RestConfig)
	// if err != nil {
	// 	return err
	// }

	// res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}

	// desired := &unstructured.Unstructured{
	// 	Object: map[string]interface{}{
	// 		"apiVersion": "v1",
	// 		"kind":       "namespace",
	// 		"metadata": map[string]interface{}{
	// 			"generateName": runner.config.Test.Definition.Namespaces["Default"].Basename,
	// 		},
	// 	},
	// }

	// created, err := client.
	// 	Resource(res).
	// 	Namespace(namespace).
	// 	Create(context.Background(), desired, metav1.CreateOptions{})
	// if err != nil {
	// 	panic(err.Error())
	// }

	// fmt.Printf("Created ConfigMap %s/%s\n", namespace, created.GetName())

	// data, _, _ := unstructured.NestedStringMap(created.Object, "data")
	// if !reflect.DeepEqual(map[string]string{"foo": "bar"}, data) {
	// 	panic("Created ConfigMap has unexpected data")
	// }

	// // Read
	// read, err := client.
	// 	Resource(res).
	// 	Namespace(namespace).
	// 	Get(
	// 		context.Background(),
	// 		created.GetName(),
	// 		metav1.GetOptions{},
	// 	)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// fmt.Printf("Read ConfigMap %s/%s\n", namespace, read.GetName())

	// data, _, _ = unstructured.NestedStringMap(read.Object, "data")
	// if !reflect.DeepEqual(map[string]string{"foo": "bar"}, data) {
	// 	panic("Read ConfigMap has unexpected data")
	// }

	// // Update
	// unstructured.SetNestedField(read.Object, "qux", "data", "foo")
	// updated, err := client.
	// 	Resource(res).
	// 	Namespace(namespace).
	// 	Update(
	// 		context.Background(),
	// 		read,
	// 		metav1.UpdateOptions{},
	// 	)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// fmt.Printf("Updated ConfigMap %s/%s\n", namespace, updated.GetName())

	// data, _, _ = unstructured.NestedStringMap(updated.Object, "data")
	// if !reflect.DeepEqual(map[string]string{"foo": "qux"}, data) {
	// 	panic("Updated ConfigMap has unexpected data")
	// }

	// // Delete
	// err = client.
	// 	Resource(res).
	// 	Namespace(namespace).
	// 	Delete(
	// 		context.Background(),
	// 		created.GetName(),
	// 		metav1.DeleteOptions{},
	// 	)
	// if err != nil {
	// 	panic(err.Error())
	// }
	// fmt.Printf("Deleted ConfigMap %s/%s\n", namespace, created.GetName())

	return nil
}
