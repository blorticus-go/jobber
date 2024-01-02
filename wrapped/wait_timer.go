package wrapped

import (
	"context"
	"fmt"
	"time"
)

type Updatable interface {
	UpdateStatus() error
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

var ErrorTimeExceeded = fmt.Errorf("time limit exceeded")

type WaitTimerExpectationFunction func(objectToTest Updatable) (expectationReached bool, errorOccurred error)

func (t *WaitTimer) TestExpectation(againstObject Updatable, expectationFunc WaitTimerExpectationFunction) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), t.MaximumTimeToWait)
	defer cancel()

	ticker := time.NewTicker(t.ProbeInterval)

	if err := againstObject.UpdateStatus(); err != nil {
		return fmt.Errorf("could not update status: %s", err)
	}

	for {
		if expectationReached, err := expectationFunc(againstObject); expectationReached {
			return nil
		} else if err != nil {
			return err
		}

		select {
		case <-ticker.C:
			if err = againstObject.UpdateStatus(); err != nil {
				return fmt.Errorf("could not update status: %s", err)
			}

		case <-ctx.Done():
			return ErrorTimeExceeded
		}
	}
}
