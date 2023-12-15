package jobber

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type K8sApiObject interface {
	GetObjectKind() schema.ObjectKind
	GetName() string
	GetNamespace() string
}

type WaitTimer struct {
	MaximumTimeToWait time.Duration
	ProbeInterval     time.Duration
}

func NewWaitTimer(maximumTimetoWait time.Duration, probeInterval time.Duration) *WaitTimer {
	return &WaitTimer{
		MaximumTimeToWait: maximumTimetoWait,
		ProbeInterval:     probeInterval,
	}
}

var TimeExceededError = fmt.Errorf("time limit exceeded")

type WaitTimerStatusUpdateFunction func(K8sApiObject) (updated K8sApiObject, err error)
type WaitTimerExpectationFunction func(K8sApiObject) (expectationReached bool, errorOccurred error)

func (t *WaitTimer) TestExpectation(againstApiObject K8sApiObject, statusUpdateFunc WaitTimerStatusUpdateFunction, expectationFunc WaitTimerExpectationFunction) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), t.MaximumTimeToWait)
	defer cancel()

	ticker := time.NewTicker(t.ProbeInterval)

	if againstApiObject, err := statusUpdateFunc(againstApiObject); err != nil {
		return fmt.Errorf("could not update status for %s (%s): %s", againstApiObject.GetObjectKind().GroupVersionKind().Kind, againstApiObject.GetName(), err)
	}

	for {
		if expectationReached, err := expectationFunc(againstApiObject); expectationReached {
			return nil
		} else if err != nil {
			return err
		}

		select {
		case <-ticker.C:
			if againstApiObject, err := statusUpdateFunc(againstApiObject); err != nil {
				return fmt.Errorf("could not update status for %s (%s): %s", againstApiObject.GetObjectKind().GroupVersionKind().Kind, againstApiObject.GetName(), err)
			}

		case <-ctx.Done():
			return TimeExceededError
		}
	}
}
