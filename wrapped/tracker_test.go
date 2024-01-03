package wrapped_test

import (
	"fmt"
	"testing"

	"github.com/blorticus-go/jobber/wrapped"
)

type ExpectedDeletionResults struct {
	SuccessfullyDeletedResources   []wrapped.Resource
	ResourceForWhichDeletionFailed wrapped.Resource
	ExpectAnError                  bool
}

func CompareDeletionResults(expected *ExpectedDeletionResults, got *wrapped.DeletionResult) error {
	if expected == nil {
		if got != nil {
			return fmt.Errorf("expected nil DeletionResult, got non-nil")
		}
		return nil
	} else if got == nil {
		return fmt.Errorf("did not expect nil DeletionResult, got nil")
	}

	if len(expected.SuccessfullyDeletedResources) != len(got.SuccessfullyDeletedResources) {
		return fmt.Errorf("expected %d SuccessfullyDeletedResources, got %d", len(expected.SuccessfullyDeletedResources), len(got.SuccessfullyDeletedResources))
	}

	for i := 0; i < len(expected.SuccessfullyDeletedResources); i++ {
		e := expected.SuccessfullyDeletedResources[i]
		g := got.SuccessfullyDeletedResources[i]

		if e.Name() != g.Name() || e.NamespaceName() != g.NamespaceName() {
			return fmt.Errorf("on SuccessfullyDeletedResources[%d] expected (%s/%s) got (%s/%s)", i, e.Name(), e.NamespaceName(), g.Name(), g.NamespaceName())
		}
	}

	if expected.ResourceForWhichDeletionFailed == nil {
		if got.ResourceForWhichDeletionFailed != nil {
			return fmt.Errorf("expected nil for ResourceForWhichDeletionFailed, got non-nil")
		}
	} else if got.ResourceForWhichDeletionFailed == nil {
		return fmt.Errorf("expected non-nil for ResourceForWhichDeletionFailed, got nil")
	} else {
		e := expected.ResourceForWhichDeletionFailed
		g := got.ResourceForWhichDeletionFailed

		if e.Name() != g.Name() || e.NamespaceName() != g.NamespaceName() {
			return fmt.Errorf("for ResourceForWhichDeletionFailed, expected (%s/%s), got (%s/%s)", e.Name(), e.NamespaceName(), g.Name(), g.NamespaceName())
		}
	}

	if expected.ExpectAnError {
		if got.Error == nil {
			return fmt.Errorf("expected an Error, got none")
		}
	} else if got.Error != nil {
		return fmt.Errorf("expected no Error, got (%s)", got.Error)
	}

	return nil
}

func ExpectEmptyDeletionResults(got *wrapped.DeletionResult) error {
	if got == nil {
		return fmt.Errorf("on AttemptToDeleteAllAsYetUndeletedResource() received nil")
	}

	if len(got.SuccessfullyDeletedResources) != 0 {
		return fmt.Errorf("on AttemptToDeleteAllAsYetUndeletedResources() expected SuccessfullyDeletedResource to be length 0, got %d", len(got.SuccessfullyDeletedResources))
	}

	if got.Error != nil {
		return fmt.Errorf("on AttemptToDeleteAllAsYetUndeletedResources() expected Error to be nil, got (%s)", got.Error)
	}

	if got.ResourceForWhichDeletionFailed != nil {
		return fmt.Errorf("on AttemptToDeleteAllAsYetUndeletedResources() expected ResourceForWhichDeletionFailed to be nil, was not")
	}

	return nil
}

func TestEmptyTracker(t *testing.T) {
	tracker := wrapped.NewResourceTracker()

	if tracker == nil {
		t.Fatalf("NewResourceTracker() returned nil")
	}

	results := tracker.AttemptToDeleteAllAsYetUndeletedResources()

	if err := ExpectEmptyDeletionResults(results); err != nil {
		t.Error(err)
	}
}

func TestMultipassGoodTracker(t *testing.T) {
	tracker := wrapped.NewResourceTracker()

	if tracker == nil {
		t.Fatalf("NewResourceTracker() returned nil")
	}

	for _, r := range []*wrapped.MockResource{
		{R_Name: "a", R_NamespaceName: "1"},
		{R_Name: "b", R_NamespaceName: "1"},
		{R_Name: "c", R_NamespaceName: "2"},
	} {
		tracker.AddCreatedResource(r)
	}

	results := tracker.AttemptToDeleteAllAsYetUndeletedResources()

	if err := CompareDeletionResults(
		&ExpectedDeletionResults{
			SuccessfullyDeletedResources: []wrapped.Resource{
				&wrapped.MockResource{R_Name: "c", R_NamespaceName: "2"},
				&wrapped.MockResource{R_Name: "b", R_NamespaceName: "1"},
				&wrapped.MockResource{R_Name: "a", R_NamespaceName: "1"},
			},
		},
		results,
	); err != nil {
		t.Fatalf("on AttemptToDeleteAllAsYetUndeletedResources, unexpected diff = %s", err)
	}

	results = tracker.AttemptToDeleteAllAsYetUndeletedResources()

	if err := ExpectEmptyDeletionResults(results); err != nil {
		t.Fatalf("expected tracker to be empty after complete deletion, got %s", err)
	}

	for _, r := range []*wrapped.MockResource{
		{R_Name: "w", R_NamespaceName: "1"},
		{R_Name: "x", R_NamespaceName: "1"},
		{R_Name: "y", R_NamespaceName: "2"},
		{R_Name: "z", R_NamespaceName: "3"},
	} {
		tracker.AddCreatedResource(r)
	}

	results = tracker.AttemptToDeleteAllAsYetUndeletedResources()

	if err := CompareDeletionResults(
		&ExpectedDeletionResults{
			SuccessfullyDeletedResources: []wrapped.Resource{
				&wrapped.MockResource{R_Name: "z", R_NamespaceName: "3"},
				&wrapped.MockResource{R_Name: "y", R_NamespaceName: "2"},
				&wrapped.MockResource{R_Name: "x", R_NamespaceName: "1"},
				&wrapped.MockResource{R_Name: "w", R_NamespaceName: "1"},
			},
		},
		results,
	); err != nil {
		t.Fatalf("on AttemptToDeleteAllAsYetUndeletedResources, unexpected diff = %s", err)
	}

	results = tracker.AttemptToDeleteAllAsYetUndeletedResources()

	if err := ExpectEmptyDeletionResults(results); err != nil {
		t.Fatalf("expected tracker to be empty after complete deletion, got %s", err)
	}

}

func TestMultpassWithError(t *testing.T) {
	tracker := wrapped.NewResourceTracker()

	if tracker == nil {
		t.Fatalf("NewResourceTracker() returned nil")
	}

	for _, r := range []*wrapped.MockResource{
		{R_Name: "a", R_NamespaceName: "1"},
		{R_Name: "b", R_NamespaceName: "1"},
		{R_Name: "c", R_NamespaceName: "2", R_DeleteError: fmt.Errorf("delete error")},
		{R_Name: "d", R_NamespaceName: "3"},
	} {
		tracker.AddCreatedResource(r)
	}

	results := tracker.AttemptToDeleteAllAsYetUndeletedResources()

	if err := CompareDeletionResults(
		&ExpectedDeletionResults{
			SuccessfullyDeletedResources: []wrapped.Resource{
				&wrapped.MockResource{R_Name: "d", R_NamespaceName: "3"},
			},
			ResourceForWhichDeletionFailed: &wrapped.MockResource{R_Name: "c", R_NamespaceName: "2"},
			ExpectAnError:                  true,
		},
		results,
	); err != nil {
		t.Fatalf("on AttemptToDeleteAllAsYetUndeletedResources, unexpected diff = %s", err)
	}

	results = tracker.AttemptToDeleteAllAsYetUndeletedResources()

	if err := CompareDeletionResults(
		&ExpectedDeletionResults{
			SuccessfullyDeletedResources:   []wrapped.Resource{},
			ResourceForWhichDeletionFailed: &wrapped.MockResource{R_Name: "c", R_NamespaceName: "2"},
			ExpectAnError:                  true,
		},
		results,
	); err != nil {
		t.Fatalf("on AttemptToDeleteAllAsYetUndeletedResources, unexpected diff = %s", err)
	}

}
