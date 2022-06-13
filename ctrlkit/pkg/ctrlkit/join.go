package ctrlkit

import (
	"context"
	"sync"

	multierr "github.com/hashicorp/go-multierror"
	"github.com/samber/lo"
	ctrl "sigs.k8s.io/controller-runtime"
)

// joinResultAndErr joins results by the following rules:
//   * If there's an error, append the error into the global one
//   * If it requires requeue, set the requeue in the global one
//   * If it sets a requeue after, set the requeue after if the global one
//     if there's none or it's longer than the local one
func joinResultAndErr(result ctrl.Result, err error, lresult ctrl.Result, lerr error) (ctrl.Result, error) {
	if lerr != nil {
		err = multierr.Append(err, lerr)
	}
	if lresult.Requeue {
		result.Requeue = true
	}
	if lresult.RequeueAfter > 0 {
		if result.RequeueAfter == 0 || result.RequeueAfter > lresult.RequeueAfter {
			result.RequeueAfter = lresult.RequeueAfter
		}
	}
	return result, err
}

func runJoinActions(ctx context.Context, actions ...ReconcileAction) (result ctrl.Result, err error) {
	// Run actions one-by-one and join results.
	for _, act := range actions {
		lr, lerr := act.Run(ctx)
		result, err = joinResultAndErr(result, err, lr, lerr)
	}
	return
}

func runJoinActionsInParallel(ctx context.Context, actions ...ReconcileAction) (result ctrl.Result, err error) {
	lresults := make([]ctrl.Result, len(actions))
	lerrs := make([]error, len(actions))

	// Run each action in a new goroutine and organize with WaitGroup.
	wg := sync.WaitGroup{}

	for i := range actions {
		act := actions[i]
		lresult, lerr := &lresults[i], &lerrs[i]
		go func() {
			defer wg.Done()

			*lresult, *lerr = act.Run(ctx)
		}()
	}

	// Wait should set a memory barrier.
	wg.Wait()

	// Join results.
	for i := 0; i < len(actions); i++ {
		result, err = joinResultAndErr(result, err, lresults[i], lerrs[i])
	}

	return
}

// Join generates a ReconcileAction that joins all the actions (orders not guaranteed).
func Join(actions ...ReconcileAction) ReconcileAction {
	if len(actions) == 0 {
		panic("must provide actions to join")
	}

	// Simply return the first action if there's only one.
	if len(actions) == 1 {
		return actions[0]
	}

	// Shuffle the actions to add some randomness.
	actions = lo.Shuffle(actions)

	// Generate a ReconcileAction that joins the action results.
	return WrapAction(describeGroup("Join", actions...), func(ctx context.Context) (result ctrl.Result, err error) {
		return runJoinActions(ctx, actions...)
	})
}

// JoinOrdered generates a ReconcileAction that joins all the actions in the order of the give actions.
func JoinOrdered(actions ...ReconcileAction) ReconcileAction {
	if len(actions) == 0 {
		panic("must provide actions to join")
	}

	// Simply return the first action if there's only one.
	if len(actions) == 1 {
		return actions[0]
	}

	// Generate a ReconcileAction that joins the action results.
	return WrapAction(describeGroup("JoinOrdered", actions...), func(ctx context.Context) (result ctrl.Result, err error) {
		return runJoinActions(ctx, actions...)
	})
}

// JoinOrdered generates a ReconcileAction that joins all the actions in parallel.
func JoinInParallel(actions ...ReconcileAction) ReconcileAction {
	if len(actions) == 0 {
		panic("must provide actions to join")
	}

	// Simply return the first action if there's only one.
	if len(actions) == 1 {
		return actions[0]
	}

	// Generate a ReconcileAction that joins the action results in parallel.
	return WrapAction(describeGroup("JoinInParallel", actions...), func(ctx context.Context) (result ctrl.Result, err error) {
		return runJoinActionsInParallel(ctx, actions...)
	})
}
