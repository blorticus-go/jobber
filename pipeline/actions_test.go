package pipeline_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/blorticus-go/jobber/pipeline"
	"github.com/blorticus-go/jobber/wrapped"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var templateResource01 = ``
var templateResource02 = `---
`
var templateResource03 = `foo`
var templateResource04 = `---
apiVersion: v1
kind: Pod
metadata:
  name: nginx-producer
  labels:
    testRole: producer
spec:
  restartPolicy: Never
  containers:
    - name: producer
      image: f5vwells/cgam-perf-test-nginx:0.9.0
      securityContext:
        allowPrivilegeEscalation: false
        runAsNonRoot: true
        capabilities:
          drop: ["ALL"]
        seccompProfile:
          type: RuntimeDefault
`
var templateResource05 = `---
apiVersion: batch/v1
kind: Job
metadata:
  name: jmeter-consumer-job
  labels:
    testRole: {{ .Values.Global.Roles.consumer }}
spec:
  backoffLimit: 0
  template:
    spec:
      activeDeadlineSeconds: 80
      restartPolicy: Never
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            - topologyKey: "kubernetes.io/hostname"
              labelSelector:
                matchExpressions:
                  - key: testRole
                    operator: In
                    values:
                      - producer
      volumes:
        - name: shared-pvc
          persistentVolumeClaim:
            claimName: shared-pipeline-pvc
      containers:
        - name: consumer
          image: f5vwells/jmeter-http2:0.8.0
          env:
            - name: USING_SIDECAR
              value: "false"
            - name: USE_BUILTIN_SCENARIO
              value: "SingleServerTarget-PreciseTPS-StaticResponses"
            - name: __SCENARIO_VAR__producerIPorHostname
              value: "10.101.19.203"
            - name: __SCENARIO_VAR__producerPort
              value: "8080"
            - name: __SCENARIO_VAR__httpTransactionsPerSecond
              value: {{ .Values.Case.TPS | quote }}
            - name: __SCENARIO_VAR__numberOfConcurrentClientConnections
              value: "1"
            - name: __SCENARIO_VAR__testDurationInSeconds
              value: {{ .Values.Unit.TestDurationInSeconds | quote }}
          args: ["-l", "/opt/test_results/jmeter.jtl.log", "-j", "/opt/test_results/jmeter.log"] 
          volumeMounts:
            - name: shared-pvc
              mountPath: /opt/test_results
          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            capabilities:
              drop: ["ALL"]
            seccompProfile:
              type: RuntimeDefault
`

type ActionMessageComparitor struct {
	ExpectError    bool
	ExpectedType   pipeline.ActionMessageType
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

	var channelIsOpen bool
	var msg *pipeline.ActionMessage

	for expectedMsgIndex, expectedMsg := range testCase.expectedMessagesInOrder {
		msg, channelIsOpen = <-msgChan

		if msg == nil && !channelIsOpen {
			return fmt.Errorf("expected message number (%d) but channel was closed before receiving", expectedMsgIndex+1)
		}

		if msg.Type != expectedMsg.ExpectedType {
			if msg.Error != nil {
				return fmt.Errorf("on message number (%d) expected type (%s), got (%s) with error (%s)", expectedMsgIndex+1, pipeline.ActionMessageTypeToString(expectedMsg.ExpectedType), pipeline.ActionMessageTypeToString(msg.Type), msg.Error)
			}
			return fmt.Errorf("on message number (%d) expected type (%s), got (%s)", expectedMsgIndex+1, pipeline.ActionMessageTypeToString(expectedMsg.ExpectedType), pipeline.ActionMessageTypeToString(msg.Type))
		}

		if expectedMsg.ExpectError {
			if msg.Error == nil {
				return fmt.Errorf("on message number (%d) expected error, got none", expectedMsgIndex+1)
			}
		} else if msg.Error != nil {
			return fmt.Errorf("on message number (%d) expected no error, got error = (%s)", expectedMsgIndex+1, msg.Error)
		}

		if expectedMsg.ExpectResource {
			if msg.Resource == nil {
				return fmt.Errorf("on message number (%d) expected a resource, got nil", expectedMsgIndex+1)
			}
		} else if msg.Resource != nil {
			return fmt.Errorf("on message number (%d) expected no resource, got a resource", expectedMsgIndex+1)
		}
	}

	if channelIsOpen {
		if msg, channelIsOpen = <-msgChan; msg != nil || channelIsOpen {
			return fmt.Errorf("after message number (%d) expected channel closure, but it was not closed", len(testCase.expectedMessagesInOrder))
		}
	}

	return nil
}

func TestTemplatedResourceAction(t *testing.T) {
	for _, testCase := range []*TemplatedResourceActionTestCase{
		{
			testName:          "File that is empty should fail",
			templateString:    templateResource01,
			pipelineVariables: pipeline.NewVariables().SetDefaultNamespaceName("test"),
			expectedMessagesInOrder: []*ActionMessageComparitor{
				{
					ExpectedType: pipeline.TemplateExpandedSuccessfully,
				},
				{
					ExpectedType: pipeline.ResourceYamlParseFailed,
					ExpectError:  true,
				},
			},
		},
		{
			testName:          "File with empty yaml document should fail",
			templateString:    templateResource02,
			pipelineVariables: pipeline.NewVariables().SetDefaultNamespaceName("test"),
			expectedMessagesInOrder: []*ActionMessageComparitor{
				{
					ExpectedType: pipeline.TemplateExpandedSuccessfully,
				},
				{
					ExpectedType: pipeline.ResourceYamlParseFailed,
					ExpectError:  true,
				},
			},
		},
		{
			testName:          "File that is invalid yaml document should fail",
			templateString:    templateResource03,
			pipelineVariables: pipeline.NewVariables().SetDefaultNamespaceName("test"),
			expectedMessagesInOrder: []*ActionMessageComparitor{
				{
					ExpectedType: pipeline.TemplateExpandedSuccessfully,
				},
				{
					ExpectedType: pipeline.ResourceYamlParseFailed,
					ExpectError:  true,
				},
			},
		},
		{
			testName:          "File that is a single well-formed Pod resource should succeed",
			templateString:    templateResource04,
			pipelineVariables: pipeline.NewVariables().SetDefaultNamespaceName("test"),
			expectedMessagesInOrder: []*ActionMessageComparitor{
				{
					ExpectedType: pipeline.TemplateExpandedSuccessfully,
				},
				{
					ExpectedType:   pipeline.WaitingForPodRunningState,
					ExpectResource: true,
				},
				{
					ExpectedType:   pipeline.ResourceCreatedSuccessfully,
					ExpectResource: true,
				},
				{
					ExpectedType: pipeline.ActionCompletedSuccessfully,
				},
			},
		},
		{
			testName:       "File that is a single well-formed Job resource with expansions should succeed",
			templateString: templateResource05,
			pipelineVariables: pipeline.NewVariables().
				SetDefaultNamespaceName("test").
				SetGlobalValues(map[string]any{
					"Roles": map[string]any{
						"consumer": "consumer",
					},
				}).
				CopyWithAddedTestUnitValues("unit01", map[string]any{
					"TestDurationInSeconds": 600,
				}).
				CopyWithAddedTestCaseValues("case01", map[string]any{
					"TPS": 500,
				}),
			expectedMessagesInOrder: []*ActionMessageComparitor{
				{
					ExpectedType: pipeline.TemplateExpandedSuccessfully,
				},
				{
					ExpectedType:   pipeline.WaitingForJobCompletion,
					ExpectResource: true,
				},
				{
					ExpectedType:   pipeline.ResourceCreatedSuccessfully,
					ExpectResource: true,
				},
				{
					ExpectedType: pipeline.ActionCompletedSuccessfully,
				},
			},
		}} {
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
		defaultActionMechanic: pipeline.NewDefaultActionMechanic(wrapped.NewFactory(nil)),
	}
}

func (mechanic *TestActionMechanic) ExpandFileAsTemplate(filePath string, templateFuncMap template.FuncMap, pipelineVariables *pipeline.Variables) (expandedBuffer *bytes.Buffer, err error) {
	return mechanic.defaultActionMechanic.ExpandFileAsTemplate(filePath, sprig.FuncMap(), pipelineVariables)
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
