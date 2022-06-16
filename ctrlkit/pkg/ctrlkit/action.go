package ctrlkit

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

// ReconcileAction is the basic unit to form a workflow. It represents some reaction
// to the states it observes. It is recommended to follow the Single-Responsibility-Rule
// while designing a ReconcileAction.
type ReconcileAction interface {
	Description() string
	Run(ctx context.Context) (ctrl.Result, error)
}
