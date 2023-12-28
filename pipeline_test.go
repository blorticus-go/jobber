package jobber_test

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/blorticus-go/jobber"
// 	"github.com/go-test/deep"
// )

// type pipelineDescriptorTestCase struct {
// 	descriptorString       string
// 	expectAnError          bool
// 	expectedPipelineAction *jobber.PipelineAction
// }

// var pipelineActionTypeToString = map[jobber.PipelineActionType]string{
// 	jobber.TemplatedResource: "TemplateResource",
// 	jobber.Executable:        "Executable",
// 	jobber.ValuesTransform:   "ValuesTransform",
// }

// func (testCase *pipelineDescriptorTestCase) RunTest() error {
// 	action, err := jobber.PipelineActionFromStringDescriptor(testCase.descriptorString, "/opt/templates")
// 	if err != nil {
// 		if !testCase.expectAnError {
// 			return fmt.Errorf("did not expect an error, but got error = %s", err)
// 		}
// 		return nil
// 	} else if testCase.expectAnError {
// 		return fmt.Errorf("expected an error, but got no error")
// 	}

// 	if testCase.expectedPipelineAction.Type != action.Type {
// 		return fmt.Errorf("expected Type (%s), got Type (%s)", pipelineActionTypeToString[testCase.expectedPipelineAction.Type], pipelineActionTypeToString[action.Type])
// 	}

// 	if testCase.descriptorString != action.Descriptor {
// 		return fmt.Errorf("expected Descriptor (%s), got Descriptor (%s)", testCase.descriptorString, action.Descriptor)
// 	}

// 	return nil
// }

// func TestPipelineDescriptors(t *testing.T) {
// 	for testCaseIndex, testCase := range []*pipelineDescriptorTestCase{
// 		{
// 			descriptorString:       "resources/first",
// 			expectedPipelineAction: &jobber.PipelineAction{jobber.TemplatedResource, "resources/first", "/opt/templates/resources/first"},
// 		},
// 		{
// 			descriptorString:       "values-transforms/post-asm.sh",
// 			expectedPipelineAction: &jobber.PipelineAction{jobber.ValuesTransform, "values-transforms/post-asm.sh", "/opt/templates/values-transforms/post-asm.sh"},
// 		},
// 		{
// 			descriptorString:       "executables/extract-data.sh",
// 			expectedPipelineAction: &jobber.PipelineAction{jobber.Executable, "executables/extract-data.sh", "/opt/templates/executables/extract-data.sh"},
// 		},
// 		{
// 			descriptorString:       "resources/jobs/first",
// 			expectedPipelineAction: &jobber.PipelineAction{jobber.TemplatedResource, "resources/jobs/first", "/opt/templates/resources/jobs/first"},
// 		},
// 		{
// 			descriptorString:       "values-transforms/asm/post-asm.sh",
// 			expectedPipelineAction: &jobber.PipelineAction{jobber.ValuesTransform, "values-transforms/asm/post-asm.sh", "/opt/templates/values-transforms/asm/post-asm.sh"},
// 		},
// 		{
// 			descriptorString:       "executables/extractor/extract-data.sh",
// 			expectedPipelineAction: &jobber.PipelineAction{jobber.Executable, "executables/extractor/extract-data.sh", "/opt/templates/executables/extractor/extract-data.sh"},
// 		},
// 		{
// 			descriptorString: "",
// 			expectAnError:    true,
// 		},
// 		{
// 			descriptorString: "resources",
// 			expectAnError:    true,
// 		},
// 		{
// 			descriptorString: "values-transform",
// 			expectAnError:    true,
// 		},
// 		{
// 			descriptorString: "executables",
// 			expectAnError:    true,
// 		},
// 		{
// 			descriptorString: "flork/first",
// 			expectAnError:    true,
// 		},
// 	} {
// 		if err := testCase.RunTest(); err != nil {
// 			t.Errorf("on test case with index [%d]: %s", testCaseIndex, err)
// 		}
// 	}
// }

// func TestPipeline(t *testing.T) {
// 	descriptors := []string{
// 		"resources/nginx-producer.yaml",
// 		"resources/telemetry.yaml",
// 		"values-transforms/post-asm.sh",
// 		"resources/shared-pvc.yaml",
// 		"resources/jmeter-job.yaml",
// 		"resources/jtl-processor-job.yaml",
// 		"resources/container-resources-job.yaml",
// 		"resources/retrieval-pod.yaml",
// 		"executables/extract-data.sh",
// 	}

// 	pipeline, err := jobber.NewPipelineFromStringDescriptors(descriptors, "/opt/templates")

// 	if err != nil {
// 		t.Fatalf("did not expect an error, but got error = %s", err)
// 	}

// 	for i := 0; i < 2; i++ {
// 		actions := make([]*jobber.PipelineAction, len(descriptors))

// 		actions[0] = pipeline.Restart()
// 		if actions[0] == nil {
// 			t.Fatalf("pipeline with %d actions returned nil on Restart()", len(descriptors))
// 		}

// 		for i := 1; i < len(descriptors); i++ {
// 			n := pipeline.NextAction()
// 			if n == nil {
// 				t.Fatalf("pipeline with %d actions returned nil on NextAction() after %d calls", len(descriptors), i+1)
// 			}

// 			actions[i] = n
// 		}

// 		if pipeline.NextAction() != nil {
// 			t.Fatalf("on call number %d of NextAction() on pipeline with %d actions, expected nil but got a value", len(descriptors), len(descriptors))
// 		}

// 		if diff := deep.Equal(actions, []*jobber.PipelineAction{
// 			{jobber.TemplatedResource, "resources/nginx-producer.yaml", "/opt/templates/resources/nginx-producer.yaml"},
// 			{jobber.TemplatedResource, "resources/telemetry.yaml", "/opt/templates/resources/telemetry.yaml"},
// 			{jobber.ValuesTransform, "values-transforms/post-asm.sh", "/opt/templates/values-transforms/post-asm.sh"},
// 			{jobber.TemplatedResource, "resources/shared-pvc.yaml", "/opt/templates/resources/shared-pvc.yaml"},
// 			{jobber.TemplatedResource, "resources/jmeter-job.yaml", "/opt/templates/resources/jmeter-job.yaml"},
// 			{jobber.TemplatedResource, "resources/jtl-processor-job.yaml", "/opt/templates/resources/jtl-processor-job.yaml"},
// 			{jobber.TemplatedResource, "resources/container-resources-job.yaml", "/opt/templates/resources/container-resources-job.yaml"},
// 			{jobber.TemplatedResource, "resources/retrieval-pod.yaml", "/opt/templates/resources/retrieval-pod.yaml"},
// 			{jobber.Executable, "executables/extract-data.sh", "/opt/templates/executables/extract-data.sh"},
// 		}); diff != nil {
// 			t.Error(diff)
// 		}
// 	}

// 	descriptors = append(descriptors, "resources")

// 	_, err = jobber.NewPipelineFromStringDescriptors(descriptors, "/opt/templates")

// 	if err == nil {
// 		t.Fatalf("expected an error on NewPipelineFromStringDescriptors() with incorred descriptor, but got no error")
// 	}
// }

// func TestTrailingSlashOnBasePath(t *testing.T) {
// 	descriptors := []string{
// 		"resources/nginx-producer.yaml",
// 		"resources/telemetry.yaml",
// 		"values-transforms/post-asm.sh",
// 		"resources/shared-pvc.yaml",
// 		"resources/jmeter-job.yaml",
// 		"resources/jtl-processor-job.yaml",
// 		"resources/container-resources-job.yaml",
// 		"resources/retrieval-pod.yaml",
// 		"executables/extract-data.sh",
// 	}

// 	pipeline, err := jobber.NewPipelineFromStringDescriptors(descriptors, "/opt/templates/")

// 	if err != nil {
// 		t.Fatalf("did not expect an error, but got error = %s", err)
// 	}

// 	actions := make([]*jobber.PipelineAction, len(descriptors))

// 	for i := 0; i < len(descriptors); i++ {
// 		n := pipeline.NextAction()
// 		if n == nil {
// 			t.Fatalf("pipeline with %d actions returned nil on NextAction() after %d calls", len(descriptors), i+1)
// 		}

// 		actions[i] = n
// 	}

// 	if pipeline.NextAction() != nil {
// 		t.Fatalf("on call number %d of NextAction() on pipeline with %d actions, expected nil but got a value", len(descriptors), len(descriptors))
// 	}

// 	if diff := deep.Equal(actions, []*jobber.PipelineAction{
// 		{jobber.TemplatedResource, "resources/nginx-producer.yaml", "/opt/templates/resources/nginx-producer.yaml"},
// 		{jobber.TemplatedResource, "resources/telemetry.yaml", "/opt/templates/resources/telemetry.yaml"},
// 		{jobber.ValuesTransform, "values-transforms/post-asm.sh", "/opt/templates/values-transforms/post-asm.sh"},
// 		{jobber.TemplatedResource, "resources/shared-pvc.yaml", "/opt/templates/resources/shared-pvc.yaml"},
// 		{jobber.TemplatedResource, "resources/jmeter-job.yaml", "/opt/templates/resources/jmeter-job.yaml"},
// 		{jobber.TemplatedResource, "resources/jtl-processor-job.yaml", "/opt/templates/resources/jtl-processor-job.yaml"},
// 		{jobber.TemplatedResource, "resources/container-resources-job.yaml", "/opt/templates/resources/container-resources-job.yaml"},
// 		{jobber.TemplatedResource, "resources/retrieval-pod.yaml", "/opt/templates/resources/retrieval-pod.yaml"},
// 		{jobber.Executable, "executables/extract-data.sh", "/opt/templates/executables/extract-data.sh"},
// 	}); diff != nil {
// 		t.Error(diff)
// 	}

// }
