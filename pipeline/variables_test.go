package pipeline_test

import (
	"fmt"
	"testing"

	"github.com/blorticus-go/jobber/pipeline"
	"github.com/go-test/deep"
)

func TestVariables(t *testing.T) {
	v := pipeline.NewVariables()

	if diff := deep.Equal(v, &pipeline.Variables{
		Values: &pipeline.Values{
			Global: map[string]any{},
			Unit:   map[string]any{},
			Case:   map[string]any{},
		},
		Config: &pipeline.Config{},
		Runtime: &pipeline.Runtime{
			Context: &pipeline.RuntimeContext{},
		},
	}); diff != nil {
		t.Errorf("on empty NewVariables: %s", diff)
	}

	v.SetAssetArchiveFilePath("/some/path/file.tar.gz")

	if diff := deep.Equal(v, &pipeline.Variables{
		Values: &pipeline.Values{
			Global: map[string]any{},
			Unit:   map[string]any{},
			Case:   map[string]any{},
		},
		Config: &pipeline.Config{
			Archive: &pipeline.ConfigArchiveInformation{FilePath: "/some/path/file.tar.gz"},
		},
		Runtime: &pipeline.Runtime{
			Context: &pipeline.RuntimeContext{},
		},
	}); diff != nil {
		t.Errorf("after SetAssetArchiveFilePath first time: %s", diff)
	}

	v.SetDefaultNamespaceName("namespace-abcde")

	if diff := deep.Equal(v, &pipeline.Variables{
		Values: &pipeline.Values{
			Global: map[string]any{},
			Unit:   map[string]any{},
			Case:   map[string]any{},
		},
		Config: &pipeline.Config{
			Archive: &pipeline.ConfigArchiveInformation{FilePath: "/some/path/file.tar.gz"},
		},
		Runtime: &pipeline.Runtime{
			DefaultNamespace: &pipeline.DefaultNamespace{
				Name: "namespace-abcde",
			},
			Context: &pipeline.RuntimeContext{},
		},
	}); diff != nil {
		t.Errorf("after SetDefaultNamespaceName first time: %s", diff)
	}

	globalValues := map[string]any{
		"ImageVersions": map[string]any{
			"nginx_producer":  "0.8.0",
			"jmeter_consumer": "0.9.0",
		},
		"TestDurationInSeconds": 600,
	}

	v.SetGlobalValues(globalValues)

	if diff := deep.Equal(v, &pipeline.Variables{
		Values: &pipeline.Values{
			Global: globalValues,
			Unit:   map[string]any{},
			Case:   map[string]any{},
		},
		Config: &pipeline.Config{
			Archive: &pipeline.ConfigArchiveInformation{FilePath: "/some/path/file.tar.gz"},
		},
		Runtime: &pipeline.Runtime{
			DefaultNamespace: &pipeline.DefaultNamespace{
				Name: "namespace-abcde",
			},
			Context: &pipeline.RuntimeContext{},
		},
	}); diff != nil {
		t.Errorf("after SetGlobalValues first time: %s", diff)
	}

	newGlobalValues := map[string]any{
		"ImageVersions": map[string]any{
			"nginx_producer":  "1.0.0",
			"jmeter_consumer": "1.0.0",
		},
		"TestDurationInSeconds": 60,
		"PipelinePvc": map[string]any{
			"StorageRequest": "3Gi",
		},
	}

	v.SetAssetArchiveFilePath("/new/path/archive.tar.gz").SetDefaultNamespaceName("namespace-12345").SetGlobalValues(newGlobalValues)

	if diff := deep.Equal(v, &pipeline.Variables{
		Values: &pipeline.Values{
			Global: newGlobalValues,
			Unit:   map[string]any{},
			Case:   map[string]any{},
		},
		Config: &pipeline.Config{
			Archive: &pipeline.ConfigArchiveInformation{FilePath: "/new/path/archive.tar.gz"},
		},
		Runtime: &pipeline.Runtime{
			DefaultNamespace: &pipeline.DefaultNamespace{
				Name: "namespace-12345",
			},
			Context: &pipeline.RuntimeContext{},
		},
	}); diff != nil {
		t.Errorf("after SetXXX second time: %s", diff)
	}

	testUnitValues := map[string]any{
		"Sidecar": map[string]any{
			"Inject": false,
		},
		"OverrideTestTimeInSeconds": 500,
	}

	v1 := v.CopyWithAddedTestUnitValues("unit01", testUnitValues)

	if diff := deep.Equal(v1, &pipeline.Variables{
		Values: &pipeline.Values{
			Global: newGlobalValues,
			Unit:   testUnitValues,
			Case:   map[string]any{},
		},
		Config: &pipeline.Config{
			Archive: &pipeline.ConfigArchiveInformation{FilePath: "/new/path/archive.tar.gz"},
		},
		Runtime: &pipeline.Runtime{
			DefaultNamespace: &pipeline.DefaultNamespace{
				Name: "namespace-12345",
			},
			Context: &pipeline.RuntimeContext{
				CurrentUnit: &pipeline.RuntimeContextUnit{
					Name: "unit01",
				},
			},
		},
	}); diff != nil {
		t.Errorf("after CopyWithAddedTestUnitValues first time: %s", diff)
	}

	if err := compareDepthOfCopy(v, v1); err != nil {
		t.Error(err)
	}

	testCaseValues := map[string]any{
		"TPS": 500,
	}

	v2 := v1.CopyWithAddedTestCaseValues("case01", testCaseValues)

	if diff := deep.Equal(v2, &pipeline.Variables{
		Values: &pipeline.Values{
			Global: newGlobalValues,
			Unit:   testUnitValues,
			Case:   testCaseValues,
		},
		Config: &pipeline.Config{
			Archive: &pipeline.ConfigArchiveInformation{FilePath: "/new/path/archive.tar.gz"},
		},
		Runtime: &pipeline.Runtime{
			DefaultNamespace: &pipeline.DefaultNamespace{
				Name: "namespace-12345",
			},
			Context: &pipeline.RuntimeContext{
				CurrentUnit: &pipeline.RuntimeContextUnit{
					Name: "unit01",
				},
				CurrentCase: &pipeline.RuntimeContextCase{
					Name: "case01",
				},
			},
		},
	}); diff != nil {
		t.Errorf("after CopyWithAddedTestUnitValues first time: %s", diff)
	}

	if err := compareDepthOfCopy(v, v1); err != nil {
		t.Error(err)
	}

}

func compareDepthOfCopy(left *pipeline.Variables, right *pipeline.Variables) error {
	if left == right {
		return fmt.Errorf("root pointers should not be the same but are")
	}

	if left.Config == right.Config {
		return fmt.Errorf(".Config pointers should not be the same but are")
	}
	if left.Runtime == right.Runtime {
		return fmt.Errorf(".Runtime pointers should not be the same but are")
	}
	if left.Values == right.Values {
		return fmt.Errorf(".Values pointers should not be the same but are")
	}

	if left.Config.Archive == right.Config.Archive {
		return fmt.Errorf(".Config.Archive pointers should not be the same but are")
	}

	if left.Runtime.Context != nil {
		if left.Runtime.Context == right.Runtime.Context {
			return fmt.Errorf(".Runtime.Context pointers should not be the same but are")
		}
		if left.Runtime.Context.CurrentCase != nil {
			if left.Runtime.Context.CurrentCase == right.Runtime.Context.CurrentCase {
				return fmt.Errorf(".Runtime.Context.CurrentCase pointers should not be the same but are")
			}
		}
		if left.Runtime.Context.CurrentUnit != nil {
			if left.Runtime.Context.CurrentUnit == right.Runtime.Context.CurrentUnit {
				return fmt.Errorf(".Runtime.Context.CurrentUnit pointers should not be the same but are")
			}
		}
	}
	if left.Runtime.DefaultNamespace == right.Runtime.DefaultNamespace {
		return fmt.Errorf(".Runtime.DefaultNamespace pointers should not be the same but are")
	}

	return nil
}
