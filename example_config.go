package jobber

// ExampleConfiguration returns a Configuration struct that represents the example 
// YAML configuration from the README.md file
func ExampleConfiguration() *Configuration {
	return &Configuration{
		Test: &ConfigurationTest{
			AssetArchive: &ConfigurationAssetArchive{
				FilePath: "/opt/performance-testing/results/test-result.tar.gz",
			},
			DefaultNamespace: &ConfigurationDefaultNamespace{
				Basename: "asm-perftest-",
			},
			GlobalValues: map[string]any{
				"ImageVersions": map[string]any{
					"cgam_perf_test_nginx":   "0.9.0",
					"file_extractor":         "0.1.0",
					"jmeter_http2":           "0.8.0",
					"jtl_processor":          "0.5.0",
					"prometheus_collector":   "0.5.2",
				},
				"TestCaseDurationInSeconds": 600,
				"PipelinePvc": map[string]any{
					"StorageRequest": "3Gi",
				},
			},
			Pipeline: &ConfigurationPipeline{
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
			Cases: []*TestCase{
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
			Units: []*TestUnit{
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
