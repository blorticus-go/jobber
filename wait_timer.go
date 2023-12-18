package jobber

import (
	"context"
	"fmt"
	"time"

	"github.com/blorticus-go/jobber/resource"
)

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

var ErrorTimeExceeded = fmt.Errorf("time limit exceeded")

type WaitTimerExpectationFunction func(resource.Type) (expectationReached bool, errorOccurred error)

func (t *WaitTimer) TestExpectation(againstResource resource.Type, expectationFunc WaitTimerExpectationFunction) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), t.MaximumTimeToWait)
	defer cancel()

	ticker := time.NewTicker(t.ProbeInterval)

	if err := againstResource.UpdateStatus(); err != nil {
		return fmt.Errorf("could not update status for %s (%s): %s", againstResource.GroupVersionKind().Kind, againstResource.Name(), err)
	}

	for {
		if expectationReached, err := expectationFunc(againstResource); expectationReached {
			return nil
		} else if err != nil {
			return err
		}

		select {
		case <-ticker.C:
			if err = againstResource.UpdateStatus(); err != nil {
				return fmt.Errorf("could not update status for %s (%s): %s", againstResource.GroupVersionKind().Kind, againstResource.GetName(), err)
			}

		case <-ctx.Done():
			return ErrorTimeExceeded
		}
	}
}
