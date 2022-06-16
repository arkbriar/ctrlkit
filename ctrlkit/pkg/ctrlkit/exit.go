package ctrlkit

import (
	"errors"

	ctrl "sigs.k8s.io/controller-runtime"
)

// ErrExit exits the workflow early.
var ErrExit = errors.New("exit")

// Exit returns an empty result and an ErrExit.
func Exit() (ctrl.Result, error) {
	return ctrl.Result{}, ErrExit
}

// IgnoreExit keeps the result but returns a nil when err == ErrExit.
func IgnoreExit(r ctrl.Result, err error) (ctrl.Result, error) {
	// If it's ErrExit, ignore it.
	if err == ErrExit {
		return r, nil
	}

	// Otherwise, it might be nil or a multi error.
	return r, err
}
