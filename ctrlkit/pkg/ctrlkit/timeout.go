package ctrlkit

import (
	"context"
	"fmt"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

type timeoutAction struct {
	timeout time.Duration
	inner   ReconcileAction
}

func (act *timeoutAction) Description() string {
	return fmt.Sprintf("Timeout(%s, %s)", act.inner.Description(), act.timeout)
}

func (act *timeoutAction) Run(ctx context.Context) (ctrl.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, act.timeout)
	defer cancel()

	return act.inner.Run(ctx)
}

// Timeout wraps the reconcile action with a timeout.
func Timeout(timeout time.Duration, act ReconcileAction) ReconcileAction {
	return &timeoutAction{timeout: timeout, inner: act}
}
