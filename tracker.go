package jobber

type DeletableK8sResource struct {
	information    *K8sResourceInformation
	deletionMethod func(object any) error
}

type ResourceDeletionAttempt struct {
	Resource *DeletableK8sResource
	Error    error
}

type CreatedResourceTracker struct {
	notYetDeletedK8sResources []*DeletableK8sResource
}

func NewCreatedResourceTracker() *CreatedResourceTracker {
	return &CreatedResourceTracker{
		notYetDeletedK8sResources: make([]*DeletableK8sResource, 0, 10),
	}
}

func (tracker *CreatedResourceTracker) AddCreatedResource(r *DeletableK8sResource) {
	tracker.notYetDeletedK8sResources = append(tracker.notYetDeletedK8sResources, r)
}

func (tracker *CreatedResourceTracker) AttemptToDeleteAllAsYetUndeletedResources() []*ResourceDeletionAttempt {
	deletionAttempts := make([]*ResourceDeletionAttempt, 0, len(tracker.notYetDeletedK8sResources))

	for len(tracker.notYetDeletedK8sResources) > 0 {
		r := tracker.notYetDeletedK8sResources[len(tracker.notYetDeletedK8sResources)-1]
		err := r.deletionMethod(r)
		deletionAttempts = append(deletionAttempts, &ResourceDeletionAttempt{r, err})
		if err != nil {
			return deletionAttempts
		}

		tracker.notYetDeletedK8sResources = tracker.notYetDeletedK8sResources[:len(tracker.notYetDeletedK8sResources)-1]
	}

	return deletionAttempts
}
