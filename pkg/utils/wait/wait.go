package wait

import (
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

// PollImmediateUntil will poll until a stop condition is met
func PollImmediateUntil(interval time.Duration, condition wait.ConditionFunc, stopCh <-chan struct{}) error {
	done, err := condition()
	if err != nil {
		return err
	}
	if done {
		return nil
	}
	return wait.PollUntil(interval, condition, stopCh)
}
