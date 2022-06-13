package ctrlkit

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

func runSequentialActions(ctx context.Context, actions ...ReconcileAction) (ctrl.Result, error) {
	// Run actions one-by-one. If one action needs to requeue or requeue after, then the
	// control flow is broken and control is returned to the outer scope.
	for _, act := range actions {
		result, err := act.Run(ctx)
		if NeedsRequeue(result, err) {
			return result, err
		}
	}

	return NoRequeue()
}

func Sequential(actions ...ReconcileAction) ReconcileAction {
	if len(actions) == 0 {
		panic("must provide actions to join")
	}

	// Simply return the first action if there's only one.
	if len(actions) == 1 {
		return actions[0]
	}

	return WrapAction(describeGroup("Sequential", actions...), func(ctx context.Context) (ctrl.Result, error) {
		return runSequentialActions(ctx, actions...)
	})
}
