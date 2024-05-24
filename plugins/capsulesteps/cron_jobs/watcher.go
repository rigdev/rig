//nolint:revive
package cron_jobs

import (
	"context"

	apipipeline "github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"github.com/rigdev/rig/pkg/controller/plugin"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func onCronJobUpdated(
	obj client.Object,
	_ []*corev1.Event,
	_ plugin.ObjectWatcher,
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

	return status
}

func (p *Plugin) WatchObjectStatus(ctx context.Context, watcher plugin.CapsuleWatcher) error {
	return watcher.WatchPrimary(ctx, &batchv1.CronJob{}, onCronJobUpdated)
}
