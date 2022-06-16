package ctrlkit

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

type actionFunc func(context.Context) (ctrl.Result, error)

type actionWrapper struct {
	description string
	actionFunc
}

func (w *actionWrapper) Description() string {
	return w.description
}

func (w *actionWrapper) Run(ctx context.Context) (ctrl.Result, error) {
	return w.actionFunc(ctx)
}

// WrapAction wraps the given description and function into an action.
func WrapAction(description string, f actionFunc) ReconcileAction {
	return &actionWrapper{description: description, actionFunc: f}
}
