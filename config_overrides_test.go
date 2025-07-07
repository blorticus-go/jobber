package jobber_test

import (
	"testing"

	"github.com/blorticus-go/jobber"
	"github.com/go-test/deep"
)

func GenerateTestConfiguration() *jobber.Configuration {
	return &jobber.Configuration{
		Test: &jobber.ConfigurationTest{
			AssetArchive: &jobber.ConfigurationAssetArchive{
				FilePath: "/opt/performance-testing/results/test-result.tar.gz",
			},
			DefaultNamespace: &jobber.ConfigurationDefaultNamespace{
				Basename: "asm-perftest-",
			},
			GlobalValues: map[string]any{
				"ImageVersions": map[string]any{
					"cgam_perf_test_nginx": "0.9.0",
					"file_extractor":       "0.1.0",
					"jmeter_http2":         "0.8.0",
					"jtl_processor":        "0.5.0",
					"prometheus_collector": "0.5.2",
				},
				"TestCaseDurationInSeconds": 600,
				"PipelinePvc": map[string]any{
					"StorageRequest": "3Gi",
				},
			},
			Pipeline: &jobber.ConfigurationPipeline{
				ActionDefinitionsRootDirectory: "/opt/performance-testing/pipeline-actions-root",
				ExecutionEnvironment: map[string]string{
					"PATH":       "/opt/openshift/aspen/client:/usr/bin:/bin",
					"KUBECONFIG": "/opt/openshift/aspen/client/auth/kubeconfig",
				},
				ActionsInOrder: []string{
					"resources/istio-cni.yaml",
					"resources/nginx-producer.yaml",
					"resources/telemetry.yaml",
					"resources/shared-pvc.yaml",
					"resources/jmeter-job.yaml",
					"values-transforms/jmeter-post-job.sh",
					"resources/jtl-processor-job.yaml",
					"resources/prom-summary-job.yaml",
					"resources/extractor.yaml",
					"executables/extract-test-results.sh",
				},
			},
			Cases: []*jobber.TestCase{
				{
					Name: "100TPS",
					Values: map[string]any{
						"TPS":                         100,
						"ConcurrentClientConnections": 8,
						"Sidecar": map[string]any{
							"WorkerThreads": 2,
						},
					},
				},
				{
					Name: "1000TPS",
					Values: map[string]any{
						"TPS":                         1000,
						"ConcurrentClientConnections": 16,
						"Sidecar": map[string]any{
							"WorkerThreads": 4,
						},
					},
				},
			},
			Units: []*jobber.TestUnit{
				{
					Name: "NoTelemetry",
					Values: map[string]any{
						"Sidecar": map[string]any{
							"Use": map[string]any{
								"Telemetry": false,
							},
						},
					},
				},
				{
					Name: "WithTelemetry",
					Values: map[string]any{
						"Sidecar": map[string]any{
							"Use": map[string]any{
								"Telemetry": true,
							},
						},
					},
				},
			},
		},
	}
}

func TestConfigurationMergeOverrideValues(t *testing.T) {
	config := GenerateTestConfiguration()
	overrides := map[string]any{
		"Test.AssetArchive.FilePath":                             "/tmp/archive.tar.gz",
		".Test.DefaultNamespace.Basename":                        "test-asm-perftest-",
		"Test.GlobalValues.ImageVersions.cgam_perf_test_nginx":   "0.9.1",
		"Test.GlobalValues.ImageVersions.jmeter_http2":           "v1.0.0",
		"Test.GlobalValues.TestCaseDurationInSeconds":            1200,
		".Test.GlobalValues.PipelinePvc.StorageRequest":          "5Gi",
		"Test.Pipeline.ActionDefinitionsRootDirectory":           "/var/tmp/foo",
		"Test.Pipeline.ExecutionEnvironment.PATH":                "/var/tmp/bar",
		".Test.Pipeline.ExecutionEnvironment.KUBECONFIG":         "",
		"Test.Pipeline.ActionsInOrder.[0]":                       "resources/istio-cni-revised.yaml",
		".Test.Pipeline.ActionsInOrder.[9]":                      "resources/something",
		"Test.Pipeline.ActionsInOrder.[10]":                      "executables/foo.sh",
		".Test.Cases.[100TPS].Values.TPS":                        200,
		"Test.Cases.[100TPS].Values.Sidecar.WorkerThreads":       3,
		"Test.Cases.[1000TPS].Values.TPS":                        2000,
		".Test.Units.[NoTelemetry].Values.Sidecar.Use.Telemetry": true,
	}

	expectedConfig := GenerateTestConfiguration()
	expectedConfig.Test.AssetArchive.FilePath = "/tmp/archive.tar.gz"
	expectedConfig.Test.DefaultNamespace.Basename = "test-asm-perftest-"
	expectedConfig.Test.GlobalValues["ImageVersions"].(map[string]any)["cgam_perf_test_nginx"] = "0.9.1"
	expectedConfig.Test.GlobalValues["ImageVersions"].(map[string]any)["jmeter_http2"] = "v1.0.0"
	expectedConfig.Test.GlobalValues["TestCaseDurationInSeconds"] = 1200
	expectedConfig.Test.GlobalValues["PipelinePvc"].(map[string]any)["StorageRequest"] = "5Gi"
	expectedConfig.Test.Pipeline.ActionDefinitionsRootDirectory = "/var/tmp/foo"
	expectedConfig.Test.Pipeline.ExecutionEnvironment["PATH"] = "/var/tmp/bar"
	expectedConfig.Test.Pipeline.ExecutionEnvironment["KUBECONFIG"] = ""
	expectedConfig.Test.Pipeline.ActionsInOrder[0] = "resources/istio-cni-revised.yaml"
	expectedConfig.Test.Pipeline.ActionsInOrder[9] = "resources/something"
	expectedConfig.Test.Pipeline.ActionsInOrder = append(expectedConfig.Test.Pipeline.ActionsInOrder, "executables/foo.sh")
	expectedConfig.Test.Cases[0].Values["TPS"] = 200
	expectedConfig.Test.Cases[0].Values["Sidecar"].(map[string]any)["WorkerThreads"] = 3
	expectedConfig.Test.Cases[1].Values["TPS"] = 2000
	expectedConfig.Test.Units[0].Values["Sidecar"].(map[string]any)["Use"].(map[string]any)["Telemetry"] = true

	err := config.MergeOverrideValues(overrides)
	if err != nil {
		t.Fatalf("failed to merge override values: %v", err)
	}

	if diff := deep.Equal(config, expectedConfig); diff != nil {
		t.Errorf("configuration mismatch after merging overrides: %v", diff)
	}

}
