package k8s

import (
	"context"
	"fmt"

	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/pkg/auth"
	"github.com/rigdev/rig/pkg/iterator"
	"google.golang.org/protobuf/types/known/timestamppb"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// ListInstances implements cluster.Gateway.
func (c *Client) ListInstances(
	ctx context.Context,
	capsuleID string,
) (iterator.Iterator[*capsule.Instance], uint64, error) {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return nil, 0, err
	}

	selector, err := labels.ValidatedSelectorFromSet(instanceLabels(capsuleID))
	if err != nil {
		return nil, 0, fmt.Errorf("could not create selector: %w", err)
	}

	p := iterator.NewProducer[*capsule.Instance]()

	go func() {
		defer p.Done()

		lopts := metav1.ListOptions{
			LabelSelector: selector.String(),
		}

		for {
			pl, err := c.cs.CoreV1().
				Pods(projectID.String()).
				List(ctx, lopts)
			if err != nil {
				p.Error(err)
				return
			}

			for _, pod := range pl.Items {
				instance, err := podToInstance(pod, capsuleID)
				if err != nil {
					p.Error(err)
					return
				}

				if err := p.Value(instance); err != nil {
					p.Error(err)
					return
				}
			}

			if pl.Continue == "" {
				break
			}
			lopts.Continue = pl.Continue
		}
	}()

	return p, 0, nil
}

// RestartInstance implements cluster.Gateway.
func (c *Client) RestartInstance(ctx context.Context, capsuleID string, instanceID string) error {
	projectID, err := auth.GetProjectID(ctx)
	if err != nil {
		return err
	}

	err = c.cs.CoreV1().
		Pods(projectID.String()).
		Delete(ctx, instanceID, metav1.DeleteOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("could not delete Pod: %w", err)
	}
	return nil
}

func podToInstance(pod v1.Pod, capsuleID string) (*capsule.Instance, error) {
	i := &capsule.Instance{
		InstanceId: pod.Name,
		BuildId:    podGetContainerImage(pod, capsuleID),
		State:      podStatusToCapsuleState(pod.Status),
		CreatedAt:  timestamppb.New(pod.ObjectMeta.CreationTimestamp.Time),
	}

	if cs := podGetContainerStatus(pod, capsuleID); cs != nil {
		i.RestartCount = uint32(cs.RestartCount)

		if cs.State.Running != nil {
			i.StartedAt = timestamppb.New(cs.State.Running.StartedAt.Time)
		}

		if cs.State.Terminated != nil {
			i.FinishedAt = timestamppb.New(cs.State.Terminated.FinishedAt.Time)
		}
	}

	return i, nil
}

func podGetContainerStatus(pod v1.Pod, container string) *v1.ContainerStatus {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name == container {
			return &cs
		}
	}
	return nil
}

func podGetContainerImage(pod v1.Pod, container string) string {
	for _, cs := range pod.Spec.Containers {
		if cs.Name == container {
			return cs.Image
		}
	}
	return ""
}

func podStatusToCapsuleState(status v1.PodStatus) capsule.State {
	for _, cs := range status.ContainerStatuses {
		if cs.State.Waiting != nil && cs.State.Waiting.Reason == "CrashLoopBackOff" {
			return capsule.State_STATE_FAILED
		}
	}

	switch status.Phase {
	case v1.PodPending:
		return capsule.State_STATE_PENDING
	case v1.PodRunning:
		return capsule.State_STATE_RUNNING
	case v1.PodSucceeded:
		return capsule.State_STATE_SUCCEEDED
	case v1.PodFailed:
		return capsule.State_STATE_FAILED
	default:
		return capsule.State_STATE_UNSPECIFIED
	}
}
