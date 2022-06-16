package ctrlkit

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

type nopAction struct{}

func (act *nopAction) Description() string {
	return "Nop"
}

func (act *nopAction) Run(ctx context.Context) (ctrl.Result, error) {
	return NoRequeue()
}

// Nop is a special action that does nothing.
var Nop = &nopAction{}
