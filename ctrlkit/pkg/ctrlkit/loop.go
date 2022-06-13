package ctrlkit

import (
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

func NeedsRequeue(result ctrl.Result, err error) bool {
	return err != nil || result.Requeue || result.RequeueAfter > 0
}

func RequeueImmediately() (ctrl.Result, error) {
	return ctrl.Result{Requeue: true}, nil
}

func RequeueAfter(after time.Duration) (ctrl.Result, error) {
	return ctrl.Result{RequeueAfter: after}, nil
}

func RequeueIfError(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}

func NoRequeue() (ctrl.Result, error) {
	return ctrl.Result{}, nil
}
