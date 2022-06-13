package ctrlkit

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

var Nop = WrapAction("Nop", func(ctx context.Context) (ctrl.Result, error) {
	return NoRequeue()
})
