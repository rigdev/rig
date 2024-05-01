// +groupName=plugins.rig.dev -- Only used for config doc generation
//
//nolint:revive
package cron_jobs

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/go-hclog"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/rigdev/rig/pkg/pipeline"
	"github.com/rigdev/rig/pkg/ptr"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Name = "rigdev.cron_jobs"
)

// Configuration for the deployment plugin
// +kubebuilder:object:root=true
type Config struct{}

type Plugin struct {
	plugin.NoWatchObjectStatus

	configBytes []byte
}

func (p *Plugin) Initialize(req plugin.InitializeRequest) error {
	p.configBytes = req.Config
	return nil
}

func (p *Plugin) Run(ctx context.Context, req pipeline.CapsuleRequest, logger hclog.Logger) error {
	// We do not have any configuration for this step?
	// var config Config
	var err error
	if len(p.configBytes) > 0 {
		_, err = plugin.ParseTemplatedConfig[Config](p.configBytes, req, plugin.CapsuleStep[Config])
		if err != nil {
			return err
		}
	}

	jobs, err := p.createCronJobs(req)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		if err := req.Set(job); err != nil {
			return err
		}
	}

	return nil
}

func (p *Plugin) createCronJobs(req pipeline.CapsuleRequest) ([]*batchv1.CronJob, error) {
	var res []*batchv1.CronJob
	deployment := &appsv1.Deployment{}
	if err := req.GetNew(deployment); errors.IsNotFound(err) {
		// TODO(anders): We should support this for command jobs.
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	for _, job := range req.Capsule().Spec.CronJobs {
		var template corev1.PodTemplateSpec
		if job.Command != nil {
			template = *deployment.Spec.Template.DeepCopy()
			c := template.Spec.Containers[0]
			c.Command = []string{job.Command.Command}
			c.Args = job.Command.Args
			template.Spec.Containers[0] = c
			template.Spec.RestartPolicy = corev1.RestartPolicyNever

		} else if job.URL != nil {
			args := []string{"-G", "--fail-with-body"}
			for k, v := range job.URL.QueryParameters {
				args = append(args, "-d", fmt.Sprintf("%v=%v", url.QueryEscape(k), url.QueryEscape(v)))
			}
			urlString := fmt.Sprintf("http://%s:%v%s", req.Capsule().Name, job.URL.Port, job.URL.Path)
			args = append(args, urlString)
			template = corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:    fmt.Sprintf("%s-%s", req.Capsule().Name, job.Name),
						Image:   "quay.io/curl/curl:latest",
						Command: []string{"curl"},
						Args:    args,
					}},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			}
		} else {
			return nil, fmt.Errorf("neither Command nor URL was set on job %s", job.Name)
		}

		annotations := pipeline.CreatePodAnnotations(req)

		j := &batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", req.Capsule().Name, job.Name),
				Namespace: req.Capsule().Namespace,
				Labels: map[string]string{
					pipeline.LabelCapsule: req.Capsule().Name,
					pipeline.LabelCron:    job.Name,
				},
				Annotations: annotations,
			},
			Spec: batchv1.CronJobSpec{
				Schedule: job.Schedule,
				JobTemplate: batchv1.JobTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: annotations,
						Labels: map[string]string{
							pipeline.LabelCapsule: req.Capsule().Name,
							pipeline.LabelCron:    job.Name,
						},
					},
					Spec: batchv1.JobSpec{
						ActiveDeadlineSeconds: ptr.Convert[uint, int64](job.TimeoutSeconds),
						BackoffLimit:          ptr.Convert[uint, int32](job.MaxRetries),
						Template:              template,
					},
				},
			},
		}
		res = append(res, j)
	}

	return res, nil
}
