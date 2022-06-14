CtrlKit -- Prototype & Demo
------

## Structure

* [ctrlkit](./ctrlkit), includes a set of core functions and an ugly implementation of CLI.
* [demo](./demo), includes partially implemented operator demo for CronJob.

## Run Demo

Demo isn't fully developed. Currently only a test could be run.

```bash
$ go test -timeout 30s -run "^Test_CronJobController_Reconcile$" demo/pkg/controller -v

=== RUN   Test_CronJobController_Reconcile
{"level":"info","msg":"TODO","cronjob":"default/example","action":"RunNextScheduledJob"}
{"level":"info","msg":"TODO","cronjob":"default/example","action":"CleanUpOldJobsExceedsHistoryLimits"}
{"level":"info","msg":"TODO","cronjob":"default/example","action":"ListActiveJobsAndUpdateStatus"}
{"level":"info","msg":"Status hasn't been changed, skip update.","cronjob":"default/example","action":"UpdateCronJobStatus"}
--- PASS: Test_CronJobController_Reconcile (0.00s)
PASS
ok      demo/pkg/controller     0.643s
```