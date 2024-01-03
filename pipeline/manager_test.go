package pipeline_test

import (
	"fmt"
	"testing"
	"text/template"

	"github.com/blorticus-go/jobber/pipeline"
)

type ManagerTestCase struct {
	TestName               string
	Descriptors            []string
	ExpectedActionTypes    []pipeline.ActionType
	ExpectAnErrorOnPrepare bool
}

func (testCase *ManagerTestCase) Execute() error {
	manager := pipeline.NewManager("/some/path", pipeline.NewActionFactory(nil, make(template.FuncMap)))

	err := manager.PrepareActionsFromStringList(testCase.Descriptors)
	if err != nil {
		if !testCase.ExpectAnErrorOnPrepare {
			return fmt.Errorf("expected no error, got error = %s", err)
		}
		return nil
	} else if testCase.ExpectAnErrorOnPrepare {
		return fmt.Errorf("expected an error, got no error")
	}

	extractedActions := make([]pipeline.Action, 0, len(testCase.ExpectedActionTypes))

	iterator := manager.ActionIterator()

	for iterator.Next() {
		if iterator.Value() == nil {
			return fmt.Errorf("on iterator.Value() got nil even though .Next() returned true")
		}
		extractedActions = append(extractedActions, iterator.Value())
	}

	if len(extractedActions) != len(testCase.ExpectedActionTypes) {
		return fmt.Errorf("expected (%d) extracted Actions, got (%d)", len(testCase.ExpectedActionTypes), len(extractedActions))
	}

	for i, expectedActionType := range testCase.ExpectedActionTypes {
		if extractedActions[i].Type() != expectedActionType {
			return fmt.Errorf("on extractedActions[%d] expected ActionType (%s), got (%s)", i, pipeline.ActionTypeToString(expectedActionType), pipeline.ActionTypeToString(extractedActions[i].Type()))
		}
	}

	return nil
}

func TestManager(t *testing.T) {
	for _, testCase := range []*ManagerTestCase{
		{
			TestName:               "empty action set should not generate an error",
			Descriptors:            []string{},
			ExpectedActionTypes:    []pipeline.ActionType{},
			ExpectAnErrorOnPrepare: false,
		},
		{
			TestName: "set of valid action descriptors should not generate an error",
			Descriptors: []string{
				"resources/template01.yaml",
				"resources/jobs/template02.yaml",
				"executables/script01.sh",
				"executables/script02.sh",
				"executables/bourne/legacy/script02.sh",
				"transforms/script03.sh",
			},
			ExpectedActionTypes: []pipeline.ActionType{
				pipeline.TemplatedResource,
				pipeline.TemplatedResource,
				pipeline.Executable,
				pipeline.Executable,
				pipeline.Executable,
				pipeline.ValuesTransform,
			},
			ExpectAnErrorOnPrepare: false,
		},
		{
			TestName: "set of valid actions with an entry with an invalid type should return an error",
			Descriptors: []string{
				"resources/template01.yaml",
				"resources/jobs/template02.yaml",
				"smurf/script01.sh",
				"executables/script02.sh",
				"executables/bourne/legacy/script02.sh",
				"transforms/script03.sh",
			},
			ExpectAnErrorOnPrepare: true,
		},
		{
			TestName: "set of valid actions with an entry with an invalid target should return an error",
			Descriptors: []string{
				"resources/template01.yaml",
				"resources/jobs/template02.yaml",
				"resources//",
				"executables/script02.sh",
				"executables/bourne/legacy/script02.sh",
				"transforms/script03.sh",
			},
			ExpectAnErrorOnPrepare: true,
		},
		{
			TestName: "set of valid actions with an entry with an empty target should return an error",
			Descriptors: []string{
				"resources/template01.yaml",
				"resources/jobs/template02.yaml",
				"executables/script02.sh",
				"executables/bourne/legacy/script02.sh",
				"transforms/",
			},
			ExpectAnErrorOnPrepare: true,
		},
		{
			TestName: "set of valid actions with an entry with no target should return an error",
			Descriptors: []string{
				"resources/template01.yaml",
				"resources",
				"executables/script02.sh",
				"executables/bourne/legacy/script02.sh",
				"transforms/script03.sh",
			},
			ExpectAnErrorOnPrepare: true,
		},
		{
			TestName: "set of valid actions with an entry with an incomplete target should return an error",
			Descriptors: []string{
				"resources/template01.yaml",
				"resources/jobs/",
				"executables/script02.sh",
				"executables/bourne/legacy/script02.sh",
				"transforms/script03.sh",
			},
			ExpectAnErrorOnPrepare: true,
		},
	} {
		if err := testCase.Execute(); err != nil {
			t.Errorf("[%s]: %s", testCase.TestName, err)
		}
	}
}
