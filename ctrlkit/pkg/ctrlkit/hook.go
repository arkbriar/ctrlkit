package ctrlkit

import (
	"context"

	"github.com/go-logr/logr"
)

// CrontollerManagerActionLifeCycleHook provides lifecycle hooks for actions.
type CrontollerManagerActionLifeCycleHook interface {
	BeforeActionRun(action string, ctx context.Context, logger logr.Logger)
	AfterActionRun(action string, ctx context.Context, logger logr.Logger)
}

// EmptyCrontollerManagerActionLifeCycleHook implements empty hooks.
type EmptyCrontollerManagerActionLifeCycleHook struct {
}

func (hook *EmptyCrontollerManagerActionLifeCycleHook) BeforeActionRun(action string, ctx context.Context, logger logr.Logger) {
}

func (hook *EmptyCrontollerManagerActionLifeCycleHook) AfterActionRun(action string, ctx context.Context, logger logr.Logger) {
}
