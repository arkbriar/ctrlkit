package ctrlkit

import (
	"errors"

	ctrl "sigs.k8s.io/controller-runtime"
)

var ErrExit = errors.New("exit")

func Exit() (ctrl.Result, error) {
	return ctrl.Result{}, ErrExit
}

func IgnoreExit(r ctrl.Result, err error) (ctrl.Result, error) {
	// If it's ErrExit, ignore it.
	if err == ErrExit {
		return r, nil
	}

	// Otherwise, it might be nil or a multi error.
	return r, err
}
