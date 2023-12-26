package wrapped

type DeletionResult struct {
	SuccessfullyDeletedResources   []Resource
	ResourceForWhichDeletionFailed Resource
	Error                          error
}

type ResourceTracker struct {
	notYetDeletedResourcesInOrderAdded []Resource
}

func NewResourceTracker() *ResourceTracker {
	return &ResourceTracker{
		notYetDeletedResourcesInOrderAdded: make([]Resource, 0, 10),
	}
}

func (t *ResourceTracker) AddCreatedResource(r Resource) *ResourceTracker {
	t.notYetDeletedResourcesInOrderAdded = append(t.notYetDeletedResourcesInOrderAdded, r)
	return t
}

func (t *ResourceTracker) popFromOrderedDeletionList() Resource {
	listLength := len(t.notYetDeletedResourcesInOrderAdded)
	if listLength > 0 {
		r := t.notYetDeletedResourcesInOrderAdded[listLength-1]
		t.notYetDeletedResourcesInOrderAdded = t.notYetDeletedResourcesInOrderAdded[:listLength-1]
		return r
	}

	return nil
}

func (t *ResourceTracker) AttemptToDeleteAllAsYetUndeletedResources() *DeletionResult {
	successfullyDeletedResource := make([]Resource, 0, len(t.notYetDeletedResourcesInOrderAdded))

	for r := t.popFromOrderedDeletionList(); r != nil; r = t.popFromOrderedDeletionList() {
		if err := r.Delete(); err != nil {
			t.AddCreatedResource(r)
			return &DeletionResult{
				SuccessfullyDeletedResources:   successfullyDeletedResource,
				ResourceForWhichDeletionFailed: r,
				Error:                          err,
			}
		}

		successfullyDeletedResource = append(successfullyDeletedResource, r)
	}

	return &DeletionResult{
		SuccessfullyDeletedResources: successfullyDeletedResource,
	}
}
