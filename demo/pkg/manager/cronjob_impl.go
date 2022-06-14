package manager

import (
	"context"
	"fmt"

	"github.com/arkbriar/ctrlkit/pkg/ctrlkit"
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	apiv1 "demo/api/v1"
)

type cronJobManager struct {
	cronJob        *apiv1.CronJob
	originalStatus *apiv1.CronJobStatus
}

func newCronJobManager(cronJob *apiv1.CronJob) cronJobManager {
	return cronJobManager{
		cronJob:        cronJob,
		originalStatus: cronJob.Status.DeepCopy(),
	}
}

func (m *cronJobManager) CronJob() *apiv1.CronJob {
	return m.cronJob
}

// IsStatusUpdated tells if the status of the CronJob managed by the manager has been changed.
func (m *cronJobManager) IsStatusUpdated() bool {
	return !equality.Semantic.DeepEqual(&m.cronJob.Status, m.originalStatus)
}

type cronJobControllerManagerImpl struct {
	ctrlkit.EmptyCrontollerManagerActionLifeCycleHook

	client  client.Client
	cronJob cronJobManager
}

func (mgr *cronJobControllerManagerImpl) UpdateCronJobStatus(ctx context.Context, logger logr.Logger) (reconcile.Result, error) {
	// Guard.
	if !mgr.cronJob.IsStatusUpdated() {
		logger.Info("Status hasn't been changed, skip update.")
		return ctrlkit.NoRequeue()
	}

	// Do update.
	logger.Info("Status has been changed, update...")
	if err := mgr.client.Status().Update(ctx, mgr.cronJob.CronJob()); err != nil {
		return ctrlkit.RequeueIfError(fmt.Errorf("unable to update status: %w", err))
	}
	logger.Info("Status updated!")
	return ctrlkit.NoRequeue()
}

func (mgr *cronJobControllerManagerImpl) ListActiveJobsAndUpdateStatus(ctx context.Context, logger logr.Logger, jobs []batchv1.Job) (ctrl.Result, error) {
	// TODO
	logger.Info("TODO")
	return ctrlkit.NoRequeue()
}

func (mgr *cronJobControllerManagerImpl) CleanUpOldJobsExceedsHistoryLimits(ctx context.Context, logger logr.Logger, jobs []batchv1.Job) (ctrl.Result, error) {
	// TODO
	logger.Info("TODO")
	return ctrlkit.NoRequeue()
}

func (mgr *cronJobControllerManagerImpl) RunNextScheduledJob(ctx context.Context, logger logr.Logger) (ctrl.Result, error) {
	// TODO
	logger.Info("TODO")
	return ctrlkit.NoRequeue()
}

func NewCronJobControllerManagerImpl(client client.Client, target *apiv1.CronJob) CronJobControllerManagerImpl {
	return &cronJobControllerManagerImpl{
		client:  client,
		cronJob: newCronJobManager(target),
	}
}
