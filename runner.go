package jobber

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
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
}

func NewRunner(config *Configuration, client *Client) *Runner {
	return &Runner{
		client: client,
		config: config,
	}
}

func (runner *Runner) CreateNamespaceFromYamlString(yamlString string) error {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, gkv, err := decode([]byte(yamlString), nil, nil)

	if err != nil {
		return err
	}

	if gkv.Kind == "Namespace" {
		ns := obj.(*corev1.Namespace)
		fmt.Printf("Namespace name = %s\n", ns.Name)
	}

	return nil
}

func (runner *Runner) defaultNamespaceResourceInformation(namespaceApiObject *corev1.Namespace) *K8sResourceInformation {
	var name string
	if namespaceApiObject != nil {
		name = namespaceApiObject.Name
	} else {
		name = runner.config.Test.Definition.Namespaces["Default"].Basename
	}

	return &K8sResourceInformation{
		Kind:          "namespace",
		Name:          name,
		NamespaceName: "",
	}
}

func (runner *Runner) createDefaultNamespace() (*Event, error) {
	defaultNamespaceApiObject, err := runner.client.CreateNamespaceUsingGeneratedName(runner.config.Test.Definition.Namespaces["Default"].Basename)

}

func (runner *Runner) RunTest(eventChannel chan<- *Event) {
	eventHandler := &eventHandler{eventChannel}

	for _, testUnit := range runner.config.Test.Units {
		eventHandler.sayThatUnitStarted(testUnit)

		if err != nil {
			eventChannel <- &Event{
				Type:           ResourceCreationFailure,
				PipelinePathId: "resources/generated/namespace/Default",
				Context:        NewEventContext(testUnit, nil),
				ResourceInformation: &ResourceEvent{
					ExpandedTemplateRetriever: nil,
					ResourceInformation:       runner.defaultNamespaceResourceInformation(defaultNamespaceApiObject),
					Error:                     err,
				},
			}
			return
		} else {
			eventHandler.sayThatResourceCreationSucceeded(runner.defaultNamespaceResourceInformation(defaultNamespaceApiObject), "resources/builtin/namespace/Default", nil, testUnit, nil)
		}

		for _, testCase := range runner.config.Test.Cases {
			eventHandler.sayThatCaseStarted(testUnit, testCase)

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
