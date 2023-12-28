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
		caseName: "base good config",
		configAsString: `---
Test:
  AssetArchive:
    FilePath: /opt/performance-test/asm/$(target-version)/$(date)/test-result.tar.gz
  DefaultNamespace:
    Basename: asm-perftest-
  GlobalValues:
    ImageVersions:
      nginx_producer: 0.8.0
    TestCaseDurationInSeconds: 600
    PipelinePvc:
      StorageRequest: 3Gi
  Pipeline:
    ActionDefinitionsRootDirectory: /home/vwells/pipeline
    ActionsInOrder:
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
      Sidecar:
        Inject: false
  - Name: mTLSOnly
    Values:
      Sidecar:
        Inject: true
        Use:
          Telemetry: false
          Pcapper: false
`,
		expectedStruct: &jobber.Configuration{
			Test: &jobber.ConfigurationTest{
				AssetArchive: &jobber.ConfigurationAssetArchive{
					FilePath: "/opt/performance-test/asm/$(target-version)/$(date)/test-result.tar.gz",
				},
				DefaultNamespace: &jobber.ConfigurationDefaultNamespace{
					Basename: "asm-perftest-",
				},
				GlobalValues: map[string]any{
					"ImageVersions": map[string]any{
						"nginx_producer": "0.8.0",
					},
					"TestCaseDurationInSeconds": 600,
					"PipelinePvc": map[string]any{
						"StorageRequest": "3Gi",
					},
				},
				Pipeline: &jobber.ConfigurationPipeline{
					ActionDefinitionsRootDirectory: "/home/vwells/pipeline",
					ActionsInOrder: []string{
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
							"Sidecar": map[string]any{
								"Inject": false,
							},
						},
					},
					{
						Name: "mTLSOnly",
						Values: map[string]any{
							"Sidecar": map[string]any{
								"Inject": true,
								"Use": map[string]any{
									"Telemetry": false,
									"Pcapper":   false,
								},
							},
						},
					},
				},
			},
		},
		expectAnError: false,
	},
	{
		caseName: "GlobalValues can be an empty map but will not be nil",
		configAsString: `---
Test:
  AssetArchive:
    FilePath: /opt/performance-test/asm/$(target-version)/$(date)/test-result.tar.gz
  DefaultNamespace:
    Basename: asm-perftest-
  Pipeline:
    ActionDefinitionsRootDirectory: /home/vwells/pipeline
    ActionsInOrder:
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
      Sidecar:
        Inject: false
  - Name: mTLSOnly
    Values:
      Sidecar:
        Inject: true
        Use:
          Telemetry: false
          Pcapper: false
`,
		expectedStruct: &jobber.Configuration{
			Test: &jobber.ConfigurationTest{
				AssetArchive: &jobber.ConfigurationAssetArchive{
					FilePath: "/opt/performance-test/asm/$(target-version)/$(date)/test-result.tar.gz",
				},
				DefaultNamespace: &jobber.ConfigurationDefaultNamespace{
					Basename: "asm-perftest-",
				},
				GlobalValues: map[string]any{},
				Pipeline: &jobber.ConfigurationPipeline{
					ActionDefinitionsRootDirectory: "/home/vwells/pipeline",
					ActionsInOrder: []string{
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
							"Sidecar": map[string]any{
								"Inject": false,
							},
						},
					},
					{
						Name: "mTLSOnly",
						Values: map[string]any{
							"Sidecar": map[string]any{
								"Inject": true,
								"Use": map[string]any{
									"Telemetry": false,
									"Pcapper":   false,
								},
							},
						},
					},
				},
			},
		},
		expectAnError: false,
	},
	{
		caseName: "test cases can have empty Values but it will not be nil",
		configAsString: `---
Test:
  AssetArchive:
    FilePath: /opt/performance-test/asm/$(target-version)/$(date)/test-result.tar.gz
  DefaultNamespace:
    Basename: asm-perftest-
  GlobalValues:
    ImageVersions:
      nginx_producer: 0.8.0
    TestCaseDurationInSeconds: 600
    PipelinePvc:
      StorageRequest: 3Gi
  Pipeline:
    ActionDefinitionsRootDirectory: /home/vwells/pipeline
    ActionsInOrder:
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
  Units:
  - Name: NoSidecar
    Values:
      Sidecar:
        Inject: false
  - Name: mTLSOnly
    Values:
      Sidecar:
        Inject: true
        Use:
          Telemetry: false
          Pcapper: false
`,
		expectedStruct: &jobber.Configuration{
			Test: &jobber.ConfigurationTest{
				AssetArchive: &jobber.ConfigurationAssetArchive{
					FilePath: "/opt/performance-test/asm/$(target-version)/$(date)/test-result.tar.gz",
				},
				DefaultNamespace: &jobber.ConfigurationDefaultNamespace{
					Basename: "asm-perftest-",
				},
				GlobalValues: map[string]any{
					"ImageVersions": map[string]any{
						"nginx_producer": "0.8.0",
					},
					"TestCaseDurationInSeconds": 600,
					"PipelinePvc": map[string]any{
						"StorageRequest": "3Gi",
					},
				},
				Pipeline: &jobber.ConfigurationPipeline{
					ActionDefinitionsRootDirectory: "/home/vwells/pipeline",
					ActionsInOrder: []string{
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
						Name:   "500TPS",
						Values: map[string]any{},
					},
				},
				Units: []*jobber.TestUnit{
					{
						Name: "NoSidecar",
						Values: map[string]any{
							"Sidecar": map[string]any{
								"Inject": false,
							},
						},
					},
					{
						Name: "mTLSOnly",
						Values: map[string]any{
							"Sidecar": map[string]any{
								"Inject": true,
								"Use": map[string]any{
									"Telemetry": false,
									"Pcapper":   false,
								},
							},
						},
					},
				},
			},
		},
		expectAnError: false,
	},
	{
		caseName: "Unit Values can be empty but will not be nil",
		configAsString: `---
Test:
  AssetArchive:
    FilePath: /opt/performance-test/asm/$(target-version)/$(date)/test-result.tar.gz
  DefaultNamespace:
    Basename: asm-perftest-
  GlobalValues:
    ImageVersions:
      nginx_producer: 0.8.0
    TestCaseDurationInSeconds: 600
    PipelinePvc:
      StorageRequest: 3Gi
  Pipeline:
    ActionDefinitionsRootDirectory: /home/vwells/pipeline
    ActionsInOrder:
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
      Sidecar:
        Inject: false
  - Name: mTLSOnly
`,
		expectedStruct: &jobber.Configuration{
			Test: &jobber.ConfigurationTest{
				AssetArchive: &jobber.ConfigurationAssetArchive{
					FilePath: "/opt/performance-test/asm/$(target-version)/$(date)/test-result.tar.gz",
				},
				DefaultNamespace: &jobber.ConfigurationDefaultNamespace{
					Basename: "asm-perftest-",
				},
				GlobalValues: map[string]any{
					"ImageVersions": map[string]any{
						"nginx_producer": "0.8.0",
					},
					"TestCaseDurationInSeconds": 600,
					"PipelinePvc": map[string]any{
						"StorageRequest": "3Gi",
					},
				},
				Pipeline: &jobber.ConfigurationPipeline{
					ActionDefinitionsRootDirectory: "/home/vwells/pipeline",
					ActionsInOrder: []string{
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
							"Sidecar": map[string]any{
								"Inject": false,
							},
						},
					},
					{
						Name:   "mTLSOnly",
						Values: map[string]any{},
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
		caseName:      "AssetArchive must be present",
		expectAnError: true,
		configAsString: `---
Test:
  DefaultNamespace:
    Basename: asm-perftest-
  GlobalValues:
    ImageVersions:
      nginx_producer: 0.8.0
    TestCaseDurationInSeconds: 600
    PipelinePvc:
      StorageRequest: 3Gi
  Pipeline:
    ActionDefinitionsRootDirectory: /home/vwells/pipeline
    ActionsInOrder:
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
      Sidecar:
        Inject: false
  - Name: mTLSOnly
    Values:
      Sidecar:
        Inject: true
        Use:
          Telemetry: false
          Pcapper: false
`,
	},
	{
		caseName:      "AssetArchive.FilePath cannot be the empty string",
		expectAnError: true,
		configAsString: `---
Test:
  AssetArchive:
    FilePath: ""
  DefaultNamespace:
    Basename: asm-perftest-
  GlobalValues:
    ImageVersions:
      nginx_producer: 0.8.0
    TestCaseDurationInSeconds: 600
    PipelinePvc:
      StorageRequest: 3Gi
  Pipeline:
    ActionDefinitionsRootDirectory: /home/vwells/pipeline
    ActionsInOrder:
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
      Sidecar:
        Inject: false
  - Name: mTLSOnly
    Values:
      Sidecar:
        Inject: true
        Use:
          Telemetry: false
          Pcapper: false
`,
	},
	{
		caseName:      "DefaultNamespace must exist",
		expectAnError: true,
		configAsString: `---
Test:
  AssetArchive:
    FilePath: /opt/performance-test/asm/$(target-version)/$(date)/test-result.tar.gz
  GlobalValues:
    ImageVersions:
      nginx_producer: 0.8.0
    TestCaseDurationInSeconds: 600
    PipelinePvc:
      StorageRequest: 3Gi
  Pipeline:
    ActionDefinitionsRootDirectory: /home/vwells/pipeline
    ActionsInOrder:
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
      Sidecar:
        Inject: false
  - Name: mTLSOnly
    Values:
      Sidecar:
        Inject: true
        Use:
          Telemetry: false
          Pcapper: false
`,
	},
	{
		caseName:      "DefaultNamespace.Basename cannot be the empty string",
		expectAnError: true,
		configAsString: `---
Test:
  AssetArchive:
    FilePath: /opt/performance-test/asm/$(target-version)/$(date)/test-result.tar.gz
  DefaultNamespace:
    Basename: ""
  GlobalValues:
    ImageVersions:
      nginx_producer: 0.8.0
    TestCaseDurationInSeconds: 600
    PipelinePvc:
      StorageRequest: 3Gi
  Pipeline:
    ActionDefinitionsRootDirectory: /home/vwells/pipeline
    ActionsInOrder:
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
      Sidecar:
        Inject: false
  - Name: mTLSOnly
    Values:
      Sidecar:
        Inject: true
        Use:
          Telemetry: false
          Pcapper: false
`,
	},
	{
		caseName:      "Pipeline must be defined",
		expectAnError: true,
		configAsString: `---
Test:
  AssetArchive:
    FilePath: /opt/performance-test/asm/$(target-version)/$(date)/test-result.tar.gz
  DefaultNamespace:
    Basename: asm-perftest-
  GlobalValues:
    ImageVersions:
      nginx_producer: 0.8.0
    TestCaseDurationInSeconds: 600
    PipelinePvc:
      StorageRequest: 3Gi
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
      Sidecar:
        Inject: false
  - Name: mTLSOnly
    Values:
      Sidecar:
        Inject: true
        Use:
          Telemetry: false
          Pcapper: false
`,
	},
	{
		caseName:      "Pipeline.ActionDefinitionsRootDirectory cannot be the empty string",
		expectAnError: true,
		configAsString: `---
Test:
  AssetArchive:
    FilePath: /opt/performance-test/asm/$(target-version)/$(date)/test-result.tar.gz
  DefaultNamespace:
    Basename: asm-perftest-
  GlobalValues:
    ImageVersions:
      nginx_producer: 0.8.0
    TestCaseDurationInSeconds: 600
    PipelinePvc:
      StorageRequest: 3Gi
  Pipeline:
    ActionDefinitionsRootDirectory: ""
    ActionsInOrder:
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
      Sidecar:
        Inject: false
  - Name: mTLSOnly
    Values:
      Sidecar:
        Inject: true
        Use:
          Telemetry: false
          Pcapper: false
`,
	},
	{
		caseName:      "Pipeline.ActionsInOrder cannot be empty",
		expectAnError: true,
		configAsString: `---
Test:
  AssetArchive:
    FilePath: /opt/performance-test/asm/$(target-version)/$(date)/test-result.tar.gz
  DefaultNamespace:
    Basename: asm-perftest-
  GlobalValues:
    ImageVersions:
      nginx_producer: 0.8.0
    TestCaseDurationInSeconds: 600
    PipelinePvc:
      StorageRequest: 3Gi
  Pipeline:
    ActionDefinitionsRootDirectory: /home/vwells/pipeline
    ActionsInOrder: []
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
      Sidecar:
        Inject: false
  - Name: mTLSOnly
    Values:
      Sidecar:
        Inject: true
        Use:
          Telemetry: false
          Pcapper: false
`,
	},
	{
		caseName:      "Cases must be defined",
		expectAnError: true,
		configAsString: `---
Test:
  AssetArchive:
    FilePath: /opt/performance-test/asm/$(target-version)/$(date)/test-result.tar.gz
  DefaultNamespace:
    Basename: asm-perftest-
  GlobalValues:
    ImageVersions:
      nginx_producer: 0.8.0
    TestCaseDurationInSeconds: 600
    PipelinePvc:
      StorageRequest: 3Gi
  Pipeline:
    ActionDefinitionsRootDirectory: /home/vwells/pipeline
    ActionsInOrder:
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
      Sidecar:
        Inject: false
  - Name: mTLSOnly
    Values:
      Sidecar:
        Inject: true
        Use:
          Telemetry: false
          Pcapper: false
`,
	},
	{
		caseName:      "Cases cannot be an empty list",
		expectAnError: true,
		configAsString: `---
Test:
  AssetArchive:
    FilePath: /opt/performance-test/asm/$(target-version)/$(date)/test-result.tar.gz
  DefaultNamespace:
    Basename: asm-perftest-
  GlobalValues:
    ImageVersions:
      nginx_producer: 0.8.0
    TestCaseDurationInSeconds: 600
    PipelinePvc:
      StorageRequest: 3Gi
  Pipeline:
    ActionDefinitionsRootDirectory: /home/vwells/pipeline
    ActionsInOrder:
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
      Sidecar:
        Inject: false
  - Name: mTLSOnly
    Values:
      Sidecar:
        Inject: true
        Use:
          Telemetry: false
          Pcapper: false
`,
	},
	{
		caseName:      "Units cannot be the empty list",
		expectAnError: true,
		configAsString: `---
Test:
  AssetArchive:
    FilePath: /opt/performance-test/asm/$(target-version)/$(date)/test-result.tar.gz
  DefaultNamespace:
    Basename: asm-perftest-
  GlobalValues:
    ImageVersions:
      nginx_producer: 0.8.0
    TestCaseDurationInSeconds: 600
    PipelinePvc:
      StorageRequest: 3Gi
  Pipeline:
    ActionDefinitionsRootDirectory: /home/vwells/pipeline
    ActionsInOrder:
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
}

func TestConfigs(t *testing.T) {
	for _, testCase := range testCases {
		if err := testCase.Run(); err != nil {
			t.Error(err)
		}
	}
}
