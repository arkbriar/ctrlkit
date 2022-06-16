package ctrlkit

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

type sequentialActions struct {
	actions []ReconcileAction
}

func (act *sequentialActions) Description() string {
	return describeGroup("Sequential", act.actions...)
}

func (act *sequentialActions) Run(ctx context.Context) (ctrl.Result, error) {
	// Run actions one-by-one. If one action needs to requeue or requeue after, then the
	// control flow is broken and control is returned to the outer scope.
	for _, act := range act.actions {
		result, err := act.Run(ctx)
		if NeedsRequeue(result, err) {
			return result, err
		}
	}

	return NoRequeue()
}

// Sequential organizes the actions into a sequential flow.
func Sequential(actions ...ReconcileAction) ReconcileAction {
	if len(actions) == 0 {
		panic("must provide actions to sequential")
	}

	// Simply return the first action if there's only one.
	if len(actions) == 1 {
		return actions[0]
	}

	return &sequentialActions{actions: actions}
}
