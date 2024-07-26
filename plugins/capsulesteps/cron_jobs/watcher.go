//nolint:revive
package cron_jobs

import (
	"context"
	"strconv"
	"time"

	apipipeline "github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/pipeline"
	"google.golang.org/protobuf/types/known/timestamppb"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func onCronJobUpdated(
	obj client.Object,
	_ []*corev1.Event,
	objWatcher plugin.ObjectWatcher,
) *apipipeline.ObjectStatusInfo {
	cronJob := obj.(*batchv1.CronJob)

	status := &apipipeline.ObjectStatusInfo{
		Properties: map[string]string{},
		PlatformStatus: []*apipipeline.PlatformObjectStatus{
			{
				Name: cronJob.Name,
				Kind: &apipipeline.PlatformObjectStatus_Cronjob{
					Cronjob: &apipipeline.CronjobStatus{
						Schedule: cronJob.Spec.Schedule,
					},
				},
			},
		},
	}

	objWatcher.WatchSecondaryByLabels(labels.Set(
		map[string]string{
			pipeline.LabelCapsule: cronJob.GetLabels()[pipeline.LabelCapsule],
			pipeline.LabelCron:    cronJob.GetLabels()[pipeline.LabelCron],
		},
	).AsSelector(), &batchv1.Job{}, onJobUpdated)

	return status
}

func onJobUpdated(
	obj client.Object,
	events []*corev1.Event,
	objectWatcher plugin.ObjectWatcher,
) *apipipeline.ObjectStatusInfo {
	job := obj.(*batchv1.Job)
	cronJob, ok := job.Labels[pipeline.LabelCron]
	if !ok {
		return nil
	}
	rolloutID, _ := strconv.ParseUint(
		job.Annotations["rig.dev/rollout"], 10, 64,
	)

	state, finishedAt := getStateAndFinishTimeFromJobExecutionStatus(job)
	var finishedAtPB *timestamppb.Timestamp
	if !finishedAt.IsZero() {
		finishedAtPB = timestamppb.New(finishedAt)
	}
	info := &apipipeline.ObjectStatusInfo{
		PlatformStatus: []*apipipeline.PlatformObjectStatus{{
			Name: job.Name,
			Kind: &apipipeline.PlatformObjectStatus_JobExecution{
				JobExecution: &apipipeline.JobExecutionStatus{
					JobName:    cronJob,
					RolloutId:  rolloutID,
					Retries:    job.Status.Failed,
					CreatedAt:  timestamppb.New(job.CreationTimestamp.Time),
					FinishedAt: finishedAtPB,
					State:      state,
				},
			},
		}},
	}

	return info
}

func getStateAndFinishTimeFromJobExecutionStatus(job *batchv1.Job) (apipipeline.JobExecutionState, time.Time) {
	if len(job.Status.Conditions) == 0 {
		return apipipeline.JobExecutionState_JOB_STATE_ONGOING, time.Time{}
	}

	for _, cond := range job.Status.Conditions {
		switch cond.Type {
		case batchv1.JobComplete:
			return apipipeline.JobExecutionState_JOB_STATE_COMPLETED, job.Status.CompletionTime.Time
		case batchv1.JobFailed:
			if cond.Reason == "DeadlineExceeded" {
				return apipipeline.JobExecutionState_JOB_STATE_TERMINATED, cond.LastTransitionTime.Time
			}
			return apipipeline.JobExecutionState_JOB_STATE_FAILED, cond.LastTransitionTime.Time
		}
	}
	return apipipeline.JobExecutionState_JOB_STATE_ONGOING, time.Time{}
}

func (p *Plugin) WatchObjectStatus(ctx context.Context, watcher plugin.CapsuleWatcher) error {
	return watcher.WatchPrimary(ctx, &batchv1.CronJob{}, onCronJobUpdated)
}
