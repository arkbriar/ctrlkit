package ctrlkit

import (
	"context"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

func describeTimeout(timeout time.Duration, act ReconcileAction) string {
	return act.Description() + ",timeout=" + timeout.String()
}

func runWithTimeout(ctx context.Context, timeout time.Duration, act ReconcileAction) (ctrl.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return act.Run(ctx)
}

func Timeout(timeout time.Duration, act ReconcileAction) ReconcileAction {
	return WrapAction(describeTimeout(timeout, act), func(ctx context.Context) (ctrl.Result, error) {
		return runWithTimeout(ctx, timeout, act)
	})
}
