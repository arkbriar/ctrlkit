package ctrlkit

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
)

type parallelAction struct {
	inner ReconcileAction
}

func (act *parallelAction) Description() string {
	return fmt.Sprintf("Parallel(%s)", act.inner.Description())
}

func (act *parallelAction) Run(ctx context.Context) (result ctrl.Result, err error) {
	done := make(chan bool)
	go func() {
		result, err = act.inner.Run(ctx)
		done <- true
	}()
	<-done

	return
}

func Parallel(act ReconcileAction) ReconcileAction {
	switch act := act.(type) {
	case *parallelAction:
		return act
	default:
		return &parallelAction{inner: act}
	}
}
