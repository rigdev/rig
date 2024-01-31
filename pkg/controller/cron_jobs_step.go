package controller

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rigdev/rig/pkg/ptr"
	"golang.org/x/exp/maps"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CronJobStep struct{}

func NewCronJobStep() *CronJobStep {
	return &CronJobStep{}
}

func (s *CronJobStep) Apply(_ context.Context, req Request) error {
	jobs, err := s.createCronJobs(req)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		req.Set(req.NamedObjectKey(job.Name, _batchCronJobGVK), job)
	}

	return nil
}

func (s *CronJobStep) createCronJobs(req Request) ([]*batchv1.CronJob, error) {
	var res []*batchv1.CronJob
	deployment := Get[*appsv1.Deployment](req, req.ObjectKey(_appsDeploymentGVK))
	if deployment == nil {
		return nil, nil
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

		annotations := map[string]string{}
		maps.Copy(annotations, req.Capsule().Annotations)

		j := &batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", req.Capsule().Name, job.Name),
				Namespace: req.Capsule().Namespace,
				Labels: map[string]string{
					LabelCapsule: req.Capsule().Name,
					LabelCron:    job.Name,
				},
				Annotations: annotations,
			},
			Spec: batchv1.CronJobSpec{
				Schedule: job.Schedule,
				JobTemplate: batchv1.JobTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: annotations,
						Labels: map[string]string{
							LabelCapsule: req.Capsule().Name,
							LabelCron:    job.Name,
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
