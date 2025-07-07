package jobber_test

import (
	"strings"
	"testing"

	"github.com/blorticus-go/jobber"
	"github.com/go-test/deep"
)

func TestVariables(t *testing.T) {
	variables := jobber.NewEmptyPipelineVariables(nil)

	if diff := deep.Equal(variables, &jobber.PipelineVariables{
		Values: &jobber.PipelineVariablesValues{
			Global: map[string]any{},
			Unit:   map[string]any{},
			Case:   map[string]any{},
		},
		Context: &jobber.PipelineVariablesContext{
			TestUnitName:                         "",
			TestCaseName:                         "",
			TestCaseRetrievedAssetsDirectoryPath: "",
		},
		Runtime: &jobber.PipelineRuntimeValues{
			DefaultNamespace: &jobber.PipelineRuntimeNamespace{
				Name: "",
			},
		},
	}); diff != nil {
		t.Error(strings.Join(diff, "\t"))
	}

	variables.WithGlobalValues(map[string]any{
		"ImageVersions": map[string]any{
			"cgam_perf_test_nginx": "0.9.0",
		},
		"TestCaseDurationInSeconds": 600,
	})

	if diff := deep.Equal(variables, &jobber.PipelineVariables{
		Values: &jobber.PipelineVariablesValues{
			Global: map[string]any{
				"ImageVersions": map[string]any{
					"cgam_perf_test_nginx": "0.9.0",
				},
				"TestCaseDurationInSeconds": 600,
			},
			Unit: map[string]any{},
			Case: map[string]any{},
		},
		Context: &jobber.PipelineVariablesContext{
			TestUnitName:                         "",
			TestCaseName:                         "",
			TestCaseRetrievedAssetsDirectoryPath: "",
		},
		Runtime: &jobber.PipelineRuntimeValues{
			DefaultNamespace: &jobber.PipelineRuntimeNamespace{
				Name: "",
			},
		},
	}); diff != nil {
		t.Error(strings.Join(diff, "\t"))
	}

	variables = variables.RescopedToUnitNamed("unit01").WithUnitValues(map[string]any{
		"Sidecar": map[string]any{
			"Inject": true,
			"Use": map[string]any{
				"Telemetry": false,
				"Pcapper":   false,
			},
		},
	})

	if diff := deep.Equal(variables, &jobber.PipelineVariables{
		Values: &jobber.PipelineVariablesValues{
			Global: map[string]any{
				"ImageVersions": map[string]any{
					"cgam_perf_test_nginx": "0.9.0",
				},
				"TestCaseDurationInSeconds": 600,
			},
			Unit: map[string]any{
				"Sidecar": map[string]any{
					"Inject": true,
					"Use": map[string]any{
						"Telemetry": false,
						"Pcapper":   false,
					},
				},
			},
			Case: map[string]any{},
		},
		Context: &jobber.PipelineVariablesContext{
			TestUnitName:                         "unit01",
			TestCaseName:                         "",
			TestCaseRetrievedAssetsDirectoryPath: "",
		},
		Runtime: &jobber.PipelineRuntimeValues{
			DefaultNamespace: &jobber.PipelineRuntimeNamespace{
				Name: "",
			},
		},
	}); diff != nil {
		t.Error(strings.Join(diff, "\t"))
	}

	variables = variables.RescopedToCaseNamed("case01").WithCaseValues(map[string]any{
		"TPS":                   100,
		"ConcurrentConnections": 1,
	}).AndUsingDefaultNamespaceNamed("default-namespace").AndTestCaseRetrievedAssetsDirectoryAt("/tmp/case")

	if diff := deep.Equal(variables, &jobber.PipelineVariables{
		Values: &jobber.PipelineVariablesValues{
			Global: map[string]any{
				"ImageVersions": map[string]any{
					"cgam_perf_test_nginx": "0.9.0",
				},
				"TestCaseDurationInSeconds": 600,
			},
			Unit: map[string]any{
				"Sidecar": map[string]any{
					"Inject": true,
					"Use": map[string]any{
						"Telemetry": false,
						"Pcapper":   false,
					},
				},
			},
			Case: map[string]any{
				"TPS":                   100,
				"ConcurrentConnections": 1,
			},
		},
		Context: &jobber.PipelineVariablesContext{
			TestUnitName:                         "unit01",
			TestCaseName:                         "case01",
			TestCaseRetrievedAssetsDirectoryPath: "/tmp/case",
		},
		Runtime: &jobber.PipelineRuntimeValues{
			DefaultNamespace: &jobber.PipelineRuntimeNamespace{
				Name: "default-namespace",
			},
		},
	}); diff != nil {
		t.Error(strings.Join(diff, "\t"))
	}

}
