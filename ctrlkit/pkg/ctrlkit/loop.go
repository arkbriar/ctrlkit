package ctrlkit

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

// NeedsRequeue reports if the result and error indicates a requeue.
func NeedsRequeue(result ctrl.Result, err error) bool {
	return err != nil || result.Requeue || result.RequeueAfter > 0
}

// RequeueImmediately returns a result with requeue set to true and a nil.
func RequeueImmediately() (ctrl.Result, error) {
	return ctrl.Result{Requeue: true}, nil
}

// RequeueAfter returns a result with requeue after set to the given duration and a nil.
func RequeueAfter(after time.Duration) (ctrl.Result, error) {
	return ctrl.Result{RequeueAfter: after}, nil
}

// RequeueIfError returns an empty result with the err.
func RequeueIfError(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}

// NoRequeue returns an empty result and a nil.
func NoRequeue() (ctrl.Result, error) {
	return ctrl.Result{}, nil
}
