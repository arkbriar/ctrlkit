package ctrlkit

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

type actionFunc func(context.Context) (ctrl.Result, error)

type actionWrap struct {
	description string
	actionFunc
}

func (w *actionWrap) Description() string {
	return w.description
}

func (w *actionWrap) Run(ctx context.Context) (ctrl.Result, error) {
	return w.actionFunc(ctx)
}

func WrapAction(description string, f actionFunc) ReconcileAction {
	return &actionWrap{description: description, actionFunc: f}
}
