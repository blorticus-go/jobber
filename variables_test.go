package jobber_test

import (
	"reflect"
	"testing"

	"github.com/blorticus-go/jobber"
	"github.com/go-test/deep"
)

func TestVariablesBase(t *testing.T) {
	v := jobber.NewEmptyPipelineVariables()

	expectedStructureWhenEmpty := &jobber.PipelineVariables{
		Values: map[string]any{},
		Config: &jobber.TemplateExpansionConfigVariables{
			Namespaces: map[string]*jobber.TemplateExpansionNamespace{},
		},
		Runtime: &jobber.PipelineRuntimeValues{},
	}

	if diff := deep.Equal(v, expectedStructureWhenEmpty); diff != nil {
		t.Errorf("on NewEmptyPipelineVariables, expect empty variables, got diff = %s", diff)
	}

	c := v.DeepCopy()

	if v == c {
		t.Errorf("on DeepCopy() base pointers should not be equal, but they are")
	}
	if v.Config == c.Config {
		t.Errorf("on DeepCopy() .Config pointers should not be equal, but they are")
	}
	if reflect.ValueOf(v.Values).Pointer() == reflect.ValueOf(c.Values).Pointer() {
		t.Errorf("on DeepCopy() .Values pointers should not be equal, but they are")
	}
	if reflect.ValueOf(v.Config.Namespaces).Pointer() == reflect.ValueOf(c.Config.Namespaces).Pointer() {
		t.Errorf("on DeepCopy() .Values.Namespaces pointers should not be equal, but they are")
	}
	if diff := deep.Equal(c, expectedStructureWhenEmpty); diff != nil {
		t.Errorf("on NewEmptyPipelineVariables, expect empty variables, got diff = %s", diff)
	}

	c.AddNamespaceToConfig("Default", "some-default")

	if diff := deep.Equal(c, &jobber.PipelineVariables{
		Values: map[string]any{},
		Config: &jobber.TemplateExpansionConfigVariables{
			Namespaces: map[string]*jobber.TemplateExpansionNamespace{
				"Default": {
					GeneratedName: "some-default",
				},
			},
		},
		Runtime: &jobber.PipelineRuntimeValues{},
	}); diff != nil {
		t.Errorf("on AddNamespaceToConfig for copy, expected does not match received, diff = %s", diff)
	}

	if diff := deep.Equal(v, expectedStructureWhenEmpty); diff != nil {
		t.Errorf("on change to copy, expect empty base variables, got diff = %s", diff)
	}

	p := jobber.NewPipelineVariablesWithSeedValues(map[string]any{
		"foo": 10,
		"bar": true,
		"baz": "blah",
	})

	if diff := deep.Equal(p, &jobber.PipelineVariables{
		Values: map[string]any{
			"foo": 10,
			"bar": true,
			"baz": "blah",
		},
		Config:  &jobber.TemplateExpansionConfigVariables{Namespaces: map[string]*jobber.TemplateExpansionNamespace{}},
		Runtime: &jobber.PipelineRuntimeValues{},
	}); diff != nil {
		t.Errorf("on creation with seed values: %s", diff)
	}

	q := jobber.NewPipelineVariablesWithSeedValues(nil)

	if diff := deep.Equal(q, expectedStructureWhenEmpty); diff != nil {
		t.Errorf("on NewPipelineVariablesWithSeedValues(nil), expect empty variables, got diff = %s", diff)
	}

}

func TestVariablesMerge(t *testing.T) {
	firstMergeInVars := map[string]any{
		"TestDurationInSeconds": 600,
		"Name":                  "test",
		"Empty":                 "",
		"InjectASidecar":        true,
		"UseMtls":               true,
		"UseTelemetry":          false,
		"UsePcapper":            false,
		"Submap": map[string]string{
			"a":    "b",
			"here": "there",
		},
	}

	v := jobber.NewEmptyPipelineVariables()
	v1 := v.MergeValuesToCopy(firstMergeInVars)

	if diff := deep.Equal(v1, &jobber.PipelineVariables{
		Values: map[string]any{
			"TestDurationInSeconds": 600,
			"Name":                  "test",
			"Empty":                 "",
			"InjectASidecar":        true,
			"UseMtls":               true,
			"UseTelemetry":          false,
			"UsePcapper":            false,
			"Submap": map[string]string{
				"a":    "b",
				"here": "there",
			},
		},
		Config:  &jobber.TemplateExpansionConfigVariables{Namespaces: map[string]*jobber.TemplateExpansionNamespace{}},
		Runtime: &jobber.PipelineRuntimeValues{},
	}); diff != nil {
		t.Errorf("on first merge of variables: %s", diff)
	}

	secondMergeVars := map[string]any{
		"TestDurationInSeconds": 400,
		"Name":                  "newname",
		"TPS":                   800,
		"Verbosity":             "Maximum",
		"Submap": map[string]string{
			"a":  "baz",
			"no": "where",
		},
	}

	v2 := v1.MergeValuesToCopy(secondMergeVars)

	if diff := deep.Equal(v2, &jobber.PipelineVariables{
		Values: map[string]any{
			"TestDurationInSeconds": 400,
			"Name":                  "newname",
			"Empty":                 "",
			"InjectASidecar":        true,
			"UseMtls":               true,
			"UseTelemetry":          false,
			"UsePcapper":            false,
			"TPS":                   800,
			"Verbosity":             "Maximum",
			"Submap": map[string]string{
				"a":  "baz",
				"no": "where",
			},
		},
		Config:  &jobber.TemplateExpansionConfigVariables{Namespaces: map[string]*jobber.TemplateExpansionNamespace{}},
		Runtime: &jobber.PipelineRuntimeValues{},
	}); diff != nil {
		t.Errorf("on second merge of variables: %s", diff)
	}
}
