package jobber_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/blorticus-go/jobber"
	"github.com/go-test/deep"
)

type configTestCase struct {
	caseName       string
	configAsString string
	expectedStruct *jobber.Configuration
	expectAnError  bool
}

func (c *configTestCase) Run() error {
	config, err := jobber.ReadConfigurationYamlFromReader(strings.NewReader(c.configAsString))

	if err != nil {
		if !c.expectAnError {
			return fmt.Errorf("on test case [%s] expected no error, but got error = (%s)", c.caseName, err)
		}
		return nil
	} else if c.expectAnError {
		return fmt.Errorf("on test case [%s] expected an error, but got no error", c.caseName)
	}

	if diff := deep.Equal(config, c.expectedStruct); diff != nil {
		return fmt.Errorf("on test case [%s] generated Configuration struct does not match expected: %s", c.caseName, diff)
	}

	return nil
}

var testCases = []*configTestCase{
	{
		caseName: "goodConfig01",
		configAsString: `---
Test:
  Definition:
    Namespaces:
      Default:
        Basename: perftest
    PipelineRootDirectory: /opt/pipeline/root
    Pipeline:
      - resources/nginx-producer.yaml
      - resources/telemetry.yaml
      - values-transforms/post-asm.sh
      - resources/shared-pvc.yaml
      - resources/jmeter-job.yaml
      - resources/jtl-processor-job.yaml
      - resources/container-resources-job.yaml
      - resources/retrieval-pod.yaml
      - executables/extract-data.sh
  Cases:
    - Name: 100TPS
      Values:
        TPS: 100
    - Name: 500TPS
      Values:
        TPS: 500
  Units:
    - Name: NoSidecar
      Values:
        TestDurationInSeconds: 600
        InjectASidecar: no
    - Name: mTLSOnly
      Values:
        TestDurationInSeconds: 600
        InjectASidecar: yes
        UseMtls: yes
        UseTelemetry: no
        UsePcapper: no
`,
		expectedStruct: &jobber.Configuration{
			Test: &jobber.ConfigurationTest{
				Definition: &jobber.ConfigurationDefinition{
					Namespaces: map[string]*jobber.ConfigurationNamespace{
						"Default": {
							Basename: "perftest",
						},
					},
					PipelineRootDirectory: "/opt/pipeline/root",
					DefaultValues:         nil,
					Pipeline: []string{
						"resources/nginx-producer.yaml",
						"resources/telemetry.yaml",
						"values-transforms/post-asm.sh",
						"resources/shared-pvc.yaml",
						"resources/jmeter-job.yaml",
						"resources/jtl-processor-job.yaml",
						"resources/container-resources-job.yaml",
						"resources/retrieval-pod.yaml",
						"executables/extract-data.sh",
					},
				},
				Cases: []*jobber.TestCase{
					{
						Name: "100TPS",
						Values: map[string]any{
							"TPS": int(100),
						},
					},
					{
						Name: "500TPS",
						Values: map[string]any{
							"TPS": int(500),
						},
					},
				},
				Units: []*jobber.TestUnit{
					{
						Name: "NoSidecar",
						Values: map[string]any{
							"TestDurationInSeconds": int(600),
							"InjectASidecar":        "no",
						},
					},
					{
						Name: "mTLSOnly",
						Values: map[string]any{
							"TestDurationInSeconds": int(600),
							"InjectASidecar":        "yes",
							"UseMtls":               "yes",
							"UseTelemetry":          "no",
							"UsePcapper":            "no",
						},
					},
				},
			},
		},
		expectAnError: false,
	},
	{
		caseName:      "config cannot be empty",
		expectAnError: true,
		configAsString: `---
`,
	},
	{
		caseName:      "invalid yaml will fail",
		expectAnError: true,
		configAsString: `---
  florb
`,
	},
	{
		caseName:      ".Test must be defined",
		expectAnError: true,
		configAsString: `---
Foo:
    Definition:
      Namespaces:
        Default:
          Basename: perftestg
      PipelineRootDirectory: /opt/pipeline/root
      Pipeline:
        - resources/nginx-producer.yaml
        - resources/telemetry.yaml
        - values-transforms/post-asm.sh
        - resources/shared-pvc.yaml
        - resources/jmeter-job.yaml
        - resources/jtl-processor-job.yaml
        - resources/container-resources-job.yaml
        - resources/retrieval-pod.yaml
        - executables/extract-data.sh
    Cases:
      - Name: 100TPS
        Values:
          TPS: 100
      - Name: 500TPS
        Values:
          TPS: 500
    Units:
      - Name: NoSidecar
        Values:
          TestDurationInSeconds: 600
          InjectASidecar: no
      - Name: mTLSOnly
        Values:
          TestDurationInSeconds: 600
          InjectASidecar: yes
          UseMtls: yes
          UseTelemetry: no
          UsePcapper: no
  `,
	},
	{
		caseName:      ".Test.Definition must exist",
		expectAnError: true,
		configAsString: `---
Test:
  Cases:
    - Name: 100TPS
      Values:
        TPS: 100
    - Name: 500TPS
      Values:
        TPS: 500
  Units:
    - Name: NoSidecar
      Values:
        TestDurationInSeconds: 600
        InjectASidecar: no
    - Name: mTLSOnly
      Values:
        TestDurationInSeconds: 600
        InjectASidecar: yes
        UseMtls: yes
        UseTelemetry: no
        UsePcapper: no
`,
	},
	{
		caseName:      ".Test.Definition.Namespaces must exist",
		expectAnError: true,
		configAsString: `---
Test:
  Definition:
    PipelineRootDirectory: /opt/pipeline/root
    Pipeline:
      - resources/nginx-producer.yaml
      - resources/telemetry.yaml
      - values-transforms/post-asm.sh
      - resources/shared-pvc.yaml
      - resources/jmeter-job.yaml
      - resources/jtl-processor-job.yaml
      - resources/container-resources-job.yaml
      - resources/retrieval-pod.yaml
      - executables/extract-data.sh
  Cases:
    - Name: 100TPS
      Values:
        TPS: 100
    - Name: 500TPS
      Values:
        TPS: 500
  Units:
    - Name: NoSidecar
      Values:
        TestDurationInSeconds: 600
        InjectASidecar: no
    - Name: mTLSOnly
      Values:
        TestDurationInSeconds: 600
        InjectASidecar: yes
        UseMtls: yes
        UseTelemetry: no
        UsePcapper: no
`,
	},
	{
		caseName:      ".Test.Definition.Namespaces cannot be empty",
		expectAnError: true,
		configAsString: `---
Test:
  Definition:
    Namespaces:
    PipelineRootDirectory: /opt/pipeline/root
    Pipeline:
      - resources/nginx-producer.yaml
      - resources/telemetry.yaml
      - values-transforms/post-asm.sh
      - resources/shared-pvc.yaml
      - resources/jmeter-job.yaml
      - resources/jtl-processor-job.yaml
      - resources/container-resources-job.yaml
      - resources/retrieval-pod.yaml
      - executables/extract-data.sh
  Cases:
    - Name: 100TPS
      Values:
        TPS: 100
    - Name: 500TPS
      Values:
        TPS: 500
  Units:
    - Name: NoSidcar
      Values:
        TestDurationInSeconds: 600
        InjectASidecar: no
    - Name: mTLSOnly
      Values:
        TestDurationInSeconds: 600
        InjectASidecar: yes
        UseMtls: yes
        UseTelemetry: no
        UsePcapper: no
`,
	},
	{
		caseName:      ".Test.Definition.Namespaces.Default must exist",
		expectAnError: true,
		configAsString: `---
Test:
  Definition:
    Namespaces:
      Florp:
        Basename: florp
    PipelineRootDirectory: /opt/pipeline/root
    Pipeline:
      - resources/nginx-producer.yaml
      - resources/telemetry.yaml
      - values-transforms/post-asm.sh
      - resources/shared-pvc.yaml
      - resources/jmeter-job.yaml
      - resources/jtl-processor-job.yaml
      - resources/container-resources-job.yaml
      - resources/retrieval-pod.yaml
      - executables/extract-data.sh
  Cases:
    - Name: 100TPS
      Values:
        TPS: 100
    - Name: 500TPS
      Values:
        TPS: 500
  Units:
    - Name: NoSidecar
      Values:
        TestDurationInSeconds: 600
        InjectASidecar: no
    - Name: mTLSOnly
      Values:
        TestDurationInSeconds: 600
        InjectASidecar: yes
        UseMtls: yes
        UseTelemetry: no
        UsePcapper: no
`,
	},
	{
		caseName:      ".Test.Definition.Pipeline must exist",
		expectAnError: true,
		configAsString: `---
Test:
    Definition:
      PipelineRootDirectory: /opt/pipeline/root
      Namespaces:
        Default:
          Basename: perftest
    Cases:
      - Name: 100TPS
        Values:
          TPS: 100
      - Name: 500TPS
        Values:
          TPS: 500
    Units:
      - Name: NoSidecar
        Values:
          TestDurationInSeconds: 600
          InjectASidecar: no
      - Name: mTLSOnly
        Values:
          TestDurationInSeconds: 600
          InjectASidecar: yes
          UseMtls: yes
          UseTelemetry: no
          UsePcapper: no
  `,
	},
	{
		caseName:      ".Test.Definition.Pipeline cannot be an empty list",
		expectAnError: true,
		configAsString: `---
Test:
    Definition:
      Namespaces:
        Default:
          Basename: perftest
      PipelineRootDirectory: /opt/pipeline/root
      Pipeline: []
    Cases:
      - Name: 100TPS
        Values:
          TPS: 100
      - Name: 500TPS
        Values:
          TPS: 500
    Units:
      - Name: NoSidecar
        Values:
          TestDurationInSeconds: 600
          InjectASidecar: no
      - Name: mTLSOnly
        Values:
          TestDurationInSeconds: 600
          InjectASidecar: yes
          UseMtls: yes
          UseTelemetry: no
          UsePcapper: no
  `,
	},
	{
		caseName:      ".Test.Cases must exist",
		expectAnError: true,
		configAsString: `---
Test:
    Definition:
      Namespaces:
        Default:
          Basename: perftest
      PipelineRootDirectory: /opt/pipeline/root
      Pipeline:
        - resources/nginx-producer.yaml
        - resources/telemetry.yaml
        - values-transforms/post-asm.sh
        - resources/shared-pvc.yaml
        - resources/jmeter-job.yaml
        - resources/jtl-processor-job.yaml
        - resources/container-resources-job.yaml
        - resources/retrieval-pod.yaml
        - executables/extract-data.sh
    Units:
      - Name: NoSidecar
        Values:
          TestDurationInSeconds: 600
          InjectASidecar: no
      - Name: mTLSOnly
        Values:
          TestDurationInSeconds: 600
          InjectASidecar: yes
          UseMtls: yes
          UseTelemetry: no
          UsePcapper: no
  `,
	},
	{
		caseName:      ".Test.Cases cannot be an empty list",
		expectAnError: true,
		configAsString: `---
Test:
    Definition:
      Namespaces:
        Default:
          Basename: perftest
      PipelineRootDirectory: /opt/pipeline/root
      Pipeline:
        - resources/nginx-producer.yaml
        - resources/telemetry.yaml
        - values-transforms/post-asm.sh
        - resources/shared-pvc.yaml
        - resources/jmeter-job.yaml
        - resources/jtl-processor-job.yaml
        - resources/container-resources-job.yaml
        - resources/retrieval-pod.yaml
        - executables/extract-data.sh
    Cases: []
    Units:
      - Name: NoSidecar
        Values:
          TestDurationInSeconds: 600
          InjectASidecar: no
      - Name: mTLSOnly
        Values:
          TestDurationInSeconds: 600
          InjectASidecar: yes
          UseMtls: yes
          UseTelemetry: no
          UsePcapper: no
  `,
	},
	{
		caseName:      ".Test.Units must be defined",
		expectAnError: true,
		configAsString: `---
Test:
    Definition:
      Namespaces:
        Default:
          Basename: perftest
      PipelineRootDirectory: /opt/pipeline/root
      Pipeline:
        - resources/nginx-producer.yaml
        - resources/telemetry.yaml
        - values-transforms/post-asm.sh
        - resources/shared-pvc.yaml
        - resources/jmeter-job.yaml
        - resources/jtl-processor-job.yaml
        - resources/container-resources-job.yaml
        - resources/retrieval-pod.yaml
        - executables/extract-data.sh
    Cases:
      - Name: 100TPS
        Values:
          TPS: 100
      - Name: 500TPS
        Values:
          TPS: 500
  `,
	},
	{
		caseName:      ".Test.Units cannot be empty",
		expectAnError: true,
		configAsString: `---
Test:
    Definition:
      Namespaces:
        Default:
          Basename: perftest
      PipelineRootDirectory: /opt/pipeline/root
      Pipeline:
        - resources/nginx-producer.yaml
        - resources/telemetry.yaml
        - values-transforms/post-asm.sh
        - resources/shared-pvc.yaml
        - resources/jmeter-job.yaml
        - resources/jtl-processor-job.yaml
        - resources/container-resources-job.yaml
        - resources/retrieval-pod.yaml
        - executables/extract-data.sh
    Cases:
      - Name: 100TPS
        Values:
          TPS: 100
      - Name: 500TPS
        Values:
          TPS: 500
    Units: []
  `,
	},
	{
		caseName:      ".Test.Definition.Pipeline action type (florp) is not valid",
		expectAnError: true,
		configAsString: `---
Test:
  Definition:
    Namespaces:
      Default:
        Basename: perftest
    PipelineRootDirectory: /opt/pipeline/root
    Pipeline:
      - resources/nginx-producer.yaml
      - resources/telemetry.yaml
      - values-transforms/post-asm.sh
      - florp/shared-pvc.yaml
      - resources/jmeter-job.yaml
      - resources/jtl-processor-job.yaml
      - resources/container-resources-job.yaml
      - resources/retrieval-pod.yaml
      - executables/extract-data.sh
  Cases:
    - Name: 100TPS
      Values:
        TPS: 100
    - Name: 500TPS
      Values:
        TPS: 500
  Units:
    - Name: NoSidecar
      Values:
        TestDurationInSeconds: 600
        InjectASidecar: no
    - Name: mTLSOnly
      Values:
        TestDurationInSeconds: 600
        InjectASidecar: yes
        UseMtls: yes
        UseTelemetry: no
        UsePcapper: no
`,
	},
	{
		caseName:      ".Test.Definition.Pipeline entries must be <type>/<target>",
		expectAnError: true,
		configAsString: `---
Test:
  Definition:
    Namespaces:
      Default:
        Basename: perftest
    PipelineRootDirectory: /opt/pipeline/root
    Pipeline:
      - resources/nginx-producer.yaml
      - resources/telemetry.yaml
      - post-asm.sh
      - resources/shared-pvc.yaml
      - resources/jmeter-job.yaml
      - resources/jtl-processor-job.yaml
      - resources/container-resources-job.yaml
      - resources/retrieval-pod.yaml
      - executables/extract-data.sh
  Cases:
    - Name: 100TPS
      Values:
        TPS: 100
    - Name: 500TPS
      Values:
        TPS: 500
  Units:
    - Name: NoSidecar
      Values:
        TestDurationInSeconds: 600
        InjectASidecar: no
    - Name: mTLSOnly
      Values:
        TestDurationInSeconds: 600
        InjectASidecar: yes
        UseMtls: yes
        UseTelemetry: no
        UsePcapper: no
`,
	},
	{
		caseName:      ".Test.Definition.Pipeline entries must be <type>/<target> without additional slashes",
		expectAnError: true,
		configAsString: `---
Test:
  Definition:
    Namespaces:
      Default:
        Basename: perftest
    PipelineRootDirectory: /opt/pipeline/root
    Pipeline:
      - resources/nginx-producer.yaml
      - resources/telemetry.yaml
      - values-transforms/post-asm.sh
      - resources/shared-pvc.yaml/baz
      - resources/jmeter-job.yaml
      - resources/jtl-processor-job.yaml
      - resources/container-resources-job.yaml
      - resources/retrieval-pod.yaml
      - executables/extract-data.sh
  Cases:
    - Name: 100TPS
      Values:
        TPS: 100
    - Name: 500TPS
      Values:
        TPS: 500
  Units:
    - Name: NoSidecar
      Values:
        TestDurationInSeconds: 600
        InjectASidecar: no
    - Name: mTLSOnly
      Values:
        TestDurationInSeconds: 600
        InjectASidecar: yes
        UseMtls: yes
        UseTelemetry: no
        UsePcapper: no
`,
	},
	{
		caseName:      ".Test.Definition.PipelineDirectoryRoot must be defined",
		expectAnError: true,
		configAsString: `---
Test:
  Definition:
    Namespaces:
      Default:
        Basename: perftest
    Pipeline:
      - resources/nginx-producer.yaml
      - resources/telemetry.yaml
      - values-transforms/post-asm.sh
      - resources/shared-pvc.yaml
      - resources/jmeter-job.yaml
      - resources/jtl-processor-job.yaml
      - resources/container-resources-job.yaml
      - resources/retrieval-pod.yaml
      - executables/extract-data.sh
  Cases:
    - Name: 100TPS
      Values:
        TPS: 100
    - Name: 500TPS
      Values:
        TPS: 500
  Units:
    - Name: NoSidecar
      Values:
        TestDurationInSeconds: 600
        InjectASidecar: no
    - Name: mTLSOnly
      Values:
        TestDurationInSeconds: 600
        InjectASidecar: yes
        UseMtls: yes
        UseTelemetry: no
        UsePcapper: no
`,
	},
}

func TestConfigs(t *testing.T) {
	for _, testCase := range testCases {
		if err := testCase.Run(); err != nil {
			t.Error(err)
		}
	}
}
