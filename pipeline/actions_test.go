package pipeline_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"text/template"

	"github.com/blorticus-go/jobber/pipeline"
	"github.com/blorticus-go/jobber/wrapped"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var templateResource01 = ``
var templateResource02 = `---`
var templateResource03 = `foo`
var templateResource04 = `---
apiVersion: v1
kind: Pod
metadata:
  name: "somename"
spec:
  containers:
  - nginx
    image: f5vwells/cgam-perf-test-nginx: 0.8.0
	imagePullPolicy: IfNotPresent
	securityContext:
		seccompProfile:
			type: RuntimeDefault
`

type ActionMessageComparitor struct {
	ExpectError    bool
	ExpectType     pipeline.ActionMessageType
	ExpectResource bool
}

type TemplatedResourceActionTestCase struct {
	testName                string
	templateString          string
	pipelineVariables       *pipeline.Variables
	expectedMessagesInOrder []*ActionMessageComparitor
}

func writeTempFileWithTemplateString(templateString string) (filePath string) {
	tempFile, err := os.CreateTemp("", "jobber-unittest-*")
	if err != nil {
		panic(fmt.Sprintf("Failed to create temp file: %s", err))
	}

	if _, err := tempFile.Write([]byte(templateString)); err != nil {
		os.Remove(tempFile.Name())
		panic(fmt.Sprintf("Failed to write to temp file [%s]: %s", tempFile.Name(), err))
	}

	tempFile.Close()

	return tempFile.Name()
}

func (testCase *TemplatedResourceActionTestCase) Execute() error {
	filePath := writeTempFileWithTemplateString(testCase.templateString)
	defer os.Remove(filePath)

	action := pipeline.NewTemplatedResourceAction(filePath, make(template.FuncMap), NewTestActionMechanic())
	msgChan := make(chan *pipeline.ActionMessage)

	go action.Run(testCase.pipelineVariables, msgChan)

	for {
		msg := <-msgChan

	}
}

func TestTemplatedResourceAction(t *testing.T) {
	for _, testCase := range []*TemplatedResourceActionTestCase{
		{
			testName:                "File with empty resource should fail",
			templateString:          templateResource01,
			pipelineVariables:       pipeline.NewVariables(),
			expectedMessagesInOrder: nil,
		},
	} {
		if err := testCase.Execute(); err != nil {
			t.Errorf("[%s] %s", testCase.testName, err)
		}
	}
}

type TestActionMechanic struct {
	defaultActionMechanic *pipeline.DefaultActionMechanic
}

func NewTestActionMechanic() *TestActionMechanic {
	return &TestActionMechanic{
		defaultActionMechanic: pipeline.NewDefaultActionMechanic(nil),
	}
}

func (mechanic *TestActionMechanic) ExpandFileAsTemplate(filePath string, templateFuncMap template.FuncMap, pipelineVariables *pipeline.Variables) (expandedBuffer *bytes.Buffer, err error) {
	return mechanic.defaultActionMechanic.ExpandFileAsTemplate(filePath, templateFuncMap, pipelineVariables)
}

func (mechanic *TestActionMechanic) ProcessBytesBufferAsYamlDocuments(buff *bytes.Buffer) (nonEmptyYamlDocuments []map[string]any, err error) {
	return mechanic.defaultActionMechanic.ProcessBytesBufferAsYamlDocuments(buff)
}

func (mechanic *TestActionMechanic) ConvertDecodedYamlToResource(decodedYaml map[string]any, defaultNamespaceName string) (wrapped.Resource, error) {
	return mechanic.defaultActionMechanic.ConvertDecodedYamlToResource(decodedYaml, defaultNamespaceName)
}

func (mechanic *TestActionMechanic) CreateResource(resource wrapped.Resource) error {
	return nil
}

func (mechanic *TestActionMechanic) TreatResourceAsPodAndWaitForRunningState(r wrapped.Resource) error {
	return nil
}

func (mechanic *TestActionMechanic) TreatResourceAsAJobAndWaitForCompletion(r wrapped.Resource) error {
	return nil
}

type MockResource struct {
	name                 string
	namespaceName        string
	groupVersionKind     schema.GroupVersionKind
	groupVersionResource schema.GroupVersionResource
}

func (resource *MockResource) Name() string {
	return resource.name
}

func (resource *MockResource) NamespaceName() string {
	return resource.namespaceName
}

func (resource *MockResource) GroupVersionKind() schema.GroupVersionKind {
	return resource.groupVersionKind
}

func (resource *MockResource) GroupVersionResource() schema.GroupVersionResource {
	return resource.groupVersionResource
}

func (resource *MockResource) Create() error {
	return nil
}

func (resource *MockResource) Delete() error {
	return nil
}

func (resource *MockResource) UpdateStatus() error {
	return nil
}

func (resource *MockResource) UnstructuredApiObject() *unstructured.Unstructured {
	return nil
}

func (resource *MockResource) UnstructuredMap() map[string]any {
	return nil
}

func (resource *MockResource) IsA(gvk schema.GroupVersionKind) bool {
	myGvk := resource.groupVersionKind
	return gvk.Group == myGvk.Group && gvk.Version == myGvk.Version && gvk.Kind == myGvk.Kind
}

func (resource *MockResource) IsNotA(gvk schema.GroupVersionKind) bool {
	return !resource.IsA(gvk)
}
