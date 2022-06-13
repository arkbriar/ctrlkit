package ctrlkit

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

type ReconcileAction interface {
	Description() string
	Run(ctx context.Context) (ctrl.Result, error)
}
