package jobber_test

import (
	"strings"
	"testing"

	"github.com/blorticus-go/jobber"
	"github.com/go-test/deep"
)

var emptyConfig string = `---
`

var goodConfig01 string = `---
RunConfiguration:
    Namespaces:
        - ReferenceName: PerfTest
          BaseName: perf-test
    DefaultNamespace: PerfTest
    Pipeline:
        - resources/nginx-producer.yaml
        - resources/telemetry.yaml
        - values-transforms/post-asm.sh
        - resources/shared-pvc.yaml
        - resources/jmeter-job.yaml
        - resources/jtl-processor-job.yaml
        - resources/container-resources-job.yaml
        - resources/retrieval-pod.yaml
        - runner/extract-data.sh
`

var invalidConfig01 string = `---
RunConfiguration:
    Namespaces:
        - ReferenceName: PerfTest
    DefaultNamespace: PerfTest
    Pipeline:
        - resources/nginx-producer.yaml
        - resources/telemetry.yaml
        - values-transforms/post-asm.sh
        - resources/shared-pvc.yaml
        - resources/jmeter-job.yaml
        - resources/jtl-processor-job.yaml
        - resources/container-resources-job.yaml
        - resources/retrieval-pod.yaml
        - runner/extract-data.sh
`

func TestEmptyConfig(t *testing.T) {
	c, err := jobber.ReadConfigurationFrom(strings.NewReader(emptyConfig))
	if err != nil {
		t.Fatalf("unexpected error on ReadConfigurationFrom(emptyConfig): %s", err.Error())
	}

	if diff := deep.Equal(c, &jobber.ConfigurationYaml{}); diff != nil {
		t.Errorf("ConfigurationYaml for emptyConfig does not match expected: %s", diff)
	}
}

func TestGoodConfigs(t *testing.T) {
	c, err := jobber.ReadConfigurationFrom(strings.NewReader(goodConfig01))

	if err != nil {
		t.Fatalf("on ReadConfigurationFrom(goodConfig01) received error: %s", err.Error())
	}

	if diff := deep.Equal(c, &jobber.ConfigurationYaml{
		RunConfiguration: &jobber.RunConfigurationYaml{
			Namespaces: []*jobber.RunConfigurationNamespacesYaml{
				{
					ReferenceName: "PerfTest",
					BaseName:      "perf-test",
				},
			},
			DefaultNamespace: "PerfTest",
			Pipeline: []string{
				"resources/nginx-producer.yaml",
				"resources/telemetry.yaml",
				"values-transforms/post-asm.sh",
				"resources/shared-pvc.yaml",
				"resources/jmeter-job.yaml",
				"resources/jtl-processor-job.yaml",
				"resources/container-resources-job.yaml",
				"resources/retrieval-pod.yaml",
				"runner/extract-data.sh",
			},
		},
	}); diff != nil {
		t.Errorf("ConfigurationYaml for goodConfig01 does not match expected: %s", diff)
	}
}

func TestInvalidConfigs(t *testing.T) {
	_, err := jobber.ReadConfigurationFrom(strings.NewReader(invalidConfig01))
	if err != nil {
		t.Errorf("expected error on ReadConfigurationFrom(invalidConfig01) but received none")
	}
}
