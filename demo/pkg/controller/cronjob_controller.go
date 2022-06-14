package controller

import (
	"context"

	"github.com/arkbriar/ctrlkit/pkg/ctrlkit"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	apiv1 "demo/api/v1"
	"demo/pkg/manager"
)

type CronJobController struct {
	client.Client
	logr.Logger
}

func (c *CronJobController) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	logger := c.Logger.WithValues("cronjob", request)

	// Get CronJob object with client.
	var cronJob apiv1.CronJob
	if err := c.Client.Get(ctx, request.NamespacedName, &cronJob); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("object not found, skip")
			return ctrlkit.NoRequeue()
		}
		logger.Error(err, "unable to get object")
		return ctrlkit.RequeueIfError(err)
	}

	// Build state and impl for controller manager.
	state := manager.NewCronJobControllerManagerState(c.Client, cronJob.DeepCopy())
	impl := manager.NewCronJobControllerManagerImpl(c.Client, cronJob.DeepCopy())
	mgr := manager.NewCronJobControllerManager(state, impl, logger)

	// Always update the status after actions have run.
	defer mgr.UpdateCronJobStatus().Run(ctx)

	// Assemble the actions and run.
	return ctrlkit.IgnoreExit(
		// Run these actions and doesn't care the order, and join the results.
		ctrlkit.Join(
			// Update the status of CronJob as always.
			mgr.ListActiveJobsAndUpdateStatus(),
			// Clean the old completed/failed jobs accroding to the limits.
			mgr.CleanUpOldJobsExceedsHistoryLimits(),
			// Try to run the next scheduled job when not suspended, otherwise do nothing.
			ctrlkit.If(cronJob.Spec.Suspend == nil || *cronJob.Spec.Suspend, mgr.RunNextScheduledJob()),
		).Run(ctx),
	)
}

func (c *CronJobController) SetupWithManager(mgr ctrl.Manager) error {
	return nil
}
