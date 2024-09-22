package deployment

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	apipipeline "github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/pipeline"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/timestamppb"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func onPodUpdated(
	obj client.Object,
	events []*corev1.Event,
	watcher plugin.ObjectWatcher,
) *apipipeline.ObjectStatusInfo {
	pod := obj.(*corev1.Pod)

	rolloutID, _ := strconv.ParseUint(pod.Annotations[pipeline.RigDevRolloutLabel], 10, 64)
	status := &apipipeline.ObjectStatusInfo{
		Properties: map[string]string{},
		PlatformStatus: []*apipipeline.PlatformObjectStatus{{
			Name: pod.GetName(),
			Kind: &apipipeline.PlatformObjectStatus_Instance{
				Instance: &apipipeline.InstanceStatus{
					RolloutId: rolloutID,
					Node:      pod.Spec.NodeName,
				},
			},
		}},
	}

	makePlacementCondition(status, pod, events)

	containers := splitByContainers(status, pod, events)

	makePreparingConditions(pod, containers, watcher)
	makeRunningConditions(pod, containers, watcher)

	return status
}

func makePreparingConditions(pod *corev1.Pod, containers []containerInfo, watcher plugin.ObjectWatcher) {
	for _, c := range containers {
		makePreparingCondition(pod, c, watcher)
	}
}

func makePreparingCondition(_ *corev1.Pod, container containerInfo, _ plugin.ObjectWatcher) {
	makeImagePullingCondition(container)

	c := getObjectCondition(container.subObj.Conditions, "Pull container image")
	if c.GetState() != apipipeline.ObjectState_OBJECT_STATE_HEALTHY {
		return
	}

	if container.status.LastTerminationState != (corev1.ContainerState{}) || container.status.State.Running != nil {
		container.subObj.Conditions = append(container.subObj.Conditions, &apipipeline.ObjectCondition{
			Name:    "Preparing",
			State:   apipipeline.ObjectState_OBJECT_STATE_HEALTHY,
			Message: "Container is done starting up",
		})
		return
	}

	cond := &apipipeline.ObjectCondition{
		Name:    "Preparing",
		State:   apipipeline.ObjectState_OBJECT_STATE_PENDING,
		Message: "",
	}
	if waiting := container.status.State.Waiting; waiting != nil {
		switch waiting.Reason {
		case "CreateContainerConfigError":
			fallthrough
		case "CreateContainerError":
			cond.State = apipipeline.ObjectState_OBJECT_STATE_ERROR
			cond.Message = waiting.Message
		case "ContainerCreating":
			cond.State = apipipeline.ObjectState_OBJECT_STATE_PENDING
			cond.Message = waiting.Message
		}
	}
	container.subObj.Conditions = append(container.subObj.Conditions, cond)
}

func makeRunningConditions(pod *corev1.Pod, containers []containerInfo, watcher plugin.ObjectWatcher) {
	for _, c := range containers {
		makeRunningCondition(pod, c, watcher)
	}
}

func makeRunningCondition(pod *corev1.Pod, container containerInfo, watcher plugin.ObjectWatcher) {
	if c := getCondition(pod.Status.Conditions, "Initialized"); c != nil {
		if c.Status != v1.ConditionTrue {
			return
		}
	}

	if c := getCondition(pod.Status.Conditions, "PodReadyToStartContainers"); c != nil {
		if c.Status != v1.ConditionTrue {
			return
		}
	} else {
		// Backup, for when PodReadyToStartContainers is not supported.
		if container.status.LastTerminationState == (v1.ContainerState{}) && container.status.State.Running == nil {
			return
		}
	}

	if container.status.LastTerminationState == (corev1.ContainerState{}) && container.status.State.Running == nil {
		return
	}

	makeExecutingCondition(container)

	if container.spec.LivenessProbe != nil {
		liveness := &apipipeline.ObjectCondition{
			Name:    "Liveness",
			Message: "Waiting for instance to start",
			State:   apipipeline.ObjectState_OBJECT_STATE_PENDING,
		}

		// Add 3 seconds to allow for the probe to run and be recorded.
		probePeriod := time.Duration(container.spec.LivenessProbe.PeriodSeconds)*time.Second + 3
		if container.status.State.Running != nil {
			if time.Since(container.status.State.Running.StartedAt.Time) < probePeriod {
				liveness.Message = "Waiting for liveness check to pass"
				liveness.UpdatedAt = timestamppb.New(container.status.State.Running.StartedAt.Time)
				watcher.Reschedule(container.status.State.Running.StartedAt.Time.Add(probePeriod))
			} else {
				liveness.Message = "Instance is alive"
				liveness.State = apipipeline.ObjectState_OBJECT_STATE_HEALTHY
			}
		}

		// Find most recent event between Unhealthy Liveness and Startup probes.
		unhealthyEvent := getEventWithPrefix(container.events, "Unhealthy", "Liveness ")
		unhealthyStartupEvent := getEventWithPrefix(container.events, "Unhealthy", "Startup ")
		if unhealthyStartupEvent != nil {
			if unhealthyEvent == nil ||
				timestampFromEvent(unhealthyEvent).AsTime().Before(timestampFromEvent(unhealthyStartupEvent).AsTime()) {
				unhealthyEvent = unhealthyStartupEvent
			}
		}

		if unhealthyEvent != nil {
			ts := timestampFromEvent(unhealthyEvent)
			if container.status.State.Running != nil {
				// Check if the event is still relevant.
				if time.Since(ts.AsTime()) < probePeriod {
					liveness.Message = unhealthyEvent.Message
					liveness.UpdatedAt = ts
					liveness.State = apipipeline.ObjectState_OBJECT_STATE_PENDING
					watcher.Reschedule(ts.AsTime().Add(probePeriod))
				}
			} else {
				liveness.Message = unhealthyEvent.Message
				liveness.UpdatedAt = ts
				liveness.State = apipipeline.ObjectState_OBJECT_STATE_ERROR
			}
		}

		container.subObj.Conditions = append(container.subObj.Conditions, liveness)
	}

	if container.spec.ReadinessProbe != nil {
		ready := &apipipeline.ObjectCondition{
			Name:    "Readiness",
			Message: "Waiting for instance to accept traffic",
			State:   apipipeline.ObjectState_OBJECT_STATE_PENDING,
		}

		startedAt := pod.GetCreationTimestamp().Time
		if container.status.State.Running != nil {
			startedAt = container.status.State.Running.StartedAt.Time
		}

		readyCondition := getCondition(pod.Status.Conditions, "Ready")
		if readyCondition != nil && readyCondition.Status == v1.ConditionTrue {
			ready.Message = "Instance ready for traffic"
			ready.State = apipipeline.ObjectState_OBJECT_STATE_HEALTHY
			ready.UpdatedAt = timestamppb.New(readyCondition.LastTransitionTime.Time)
		} else {
			unhealthyEvent := getEventWithPrefix(container.events, "Unhealthy", "Readiness ")
			if unhealthyEvent != nil {
				ts := timestampFromEvent(unhealthyEvent)
				if ts.AsTime().After(startedAt) {
					ready.Message = unhealthyEvent.Message
					ready.State = apipipeline.ObjectState_OBJECT_STATE_ERROR
					ready.UpdatedAt = ts
				}
			}
		}

		container.subObj.Conditions = append(container.subObj.Conditions, ready)
	}
}

func makeExecutingCondition(container containerInfo) {
	cond := &apipipeline.ObjectCondition{
		Name:    "Running",
		Message: "Waiting to start",
		State:   apipipeline.ObjectState_OBJECT_STATE_PENDING,
	}

	container.platformStatus.RestartCount = uint32(container.status.RestartCount)

	if container.status.LastTerminationState.Terminated != nil {
		container.platformStatus.LastTermination = containerStateTerminatedFromK8s(
			container.status.LastTerminationState.Terminated)
		// If this isn't overwritten, it's because the instance is 'Waiting to start' after it had crashed
		cond.State = apipipeline.ObjectState_OBJECT_STATE_ERROR
	}

	if container.status.State.Running != nil {
		container.platformStatus.StartedAt = timestamppb.New(container.status.State.Running.StartedAt.Time)
		cond.Message = "Container is running"
		cond.State = apipipeline.ObjectState_OBJECT_STATE_HEALTHY
	} else if container.status.State.Waiting != nil {
		cond.Message = fmt.Sprintf("%s: %s", container.status.State.Waiting.Reason, container.status.State.Waiting.Message)
		cond.State = apipipeline.ObjectState_OBJECT_STATE_ERROR
	}

	container.subObj.Conditions = append(container.subObj.Conditions, cond)
}

func containerStateTerminatedFromK8s(
	state *v1.ContainerStateTerminated,
) *apipipeline.ContainerStatus_ContainerTermination {
	if state == nil {
		return nil
	}
	return &apipipeline.ContainerStatus_ContainerTermination{
		ExitCode:    state.ExitCode,
		Signal:      state.Signal,
		Reason:      state.Reason,
		Message:     state.Message,
		StartedAt:   timestamppb.New(state.StartedAt.Time),
		FinishedAt:  timestamppb.New(state.FinishedAt.Time),
		ContainerId: state.ContainerID,
	}
}

func makeImagePullingCondition(container containerInfo) {
	cond := &apipipeline.ObjectCondition{
		Name:  "Pull container image",
		State: apipipeline.ObjectState_OBJECT_STATE_PENDING,
	}

	for _, e := range container.events {
		switch {
		case e.Reason == "Pulled":
			cond.State = apipipeline.ObjectState_OBJECT_STATE_HEALTHY
		case e.Reason == "Pulling":
			cond.State = apipipeline.ObjectState_OBJECT_STATE_PENDING
		case e.Reason == "BackOff" && strings.HasPrefix(e.Message, "Back-off pulling image"):
			cond.State = apipipeline.ObjectState_OBJECT_STATE_ERROR
		case e.Reason == "Failed" && strings.HasPrefix(e.Message, "Failed to pull image"):
			cond.State = apipipeline.ObjectState_OBJECT_STATE_ERROR
		default:
			continue
		}

		cond.Message = e.Message
		cond.UpdatedAt = timestampFromEvent(e)
		break
	}

	// Override event if latest event is wrong.
	if container.status.ImageID != "" && cond.State != apipipeline.ObjectState_OBJECT_STATE_HEALTHY {
		cond.State = apipipeline.ObjectState_OBJECT_STATE_HEALTHY
		cond.Message = "Image pulled"
	}

	if cond.State != apipipeline.ObjectState_OBJECT_STATE_HEALTHY {
		// Bad state. Try to pull out a better message from the waiting status, if possible.
		if w := container.status.State.Waiting; w != nil && cond.Message == "" {
			switch w.Reason {
			case "ErrImagePull":
				cond.State = apipipeline.ObjectState_OBJECT_STATE_ERROR
			case "ImagePullBackOff":
				cond.State = apipipeline.ObjectState_OBJECT_STATE_ERROR
			default:
			}

			cond.Message = w.Message
			cond.UpdatedAt = timestamppb.Now()
		}
	}

	container.platformStatus.Image = container.status.ImageID
	container.subObj.Properties["Image"] = container.status.ImageID
	container.subObj.Conditions = append(container.subObj.Conditions, cond)
}

func makePlacementCondition(status *apipipeline.ObjectStatusInfo, pod *corev1.Pod, events []*corev1.Event) {
	cond := &apipipeline.ObjectCondition{
		Name:  "Node placement",
		State: apipipeline.ObjectState_OBJECT_STATE_PENDING,
	}

	if node := pod.Spec.NodeName; node != "" {
		status.Properties["Node"] = node
	}

	if scheduled := getCondition(pod.Status.Conditions, "PodScheduled"); scheduled != nil {
		cond.UpdatedAt = timestampFromCondition(scheduled)

		switch scheduled.Status {
		case "True":
			cond.State = apipipeline.ObjectState_OBJECT_STATE_HEALTHY
		}

		cond.Message = scheduled.Message
	}

	for _, e := range events {
		switch e.Reason {
		case "Scheduled":
			cond.State = apipipeline.ObjectState_OBJECT_STATE_HEALTHY
		case "FailedScheduling":
			cond.State = apipipeline.ObjectState_OBJECT_STATE_ERROR
		default:
			continue
		}

		cond.Message = e.Message
		cond.UpdatedAt = timestampFromEvent(e)
		break
	}

	status.Conditions = append(status.Conditions, cond)
}

type containerInfo struct {
	name           string
	status         v1.ContainerStatus
	spec           v1.Container
	events         []*v1.Event
	subObj         *apipipeline.SubObjectStatus
	platformStatus *apipipeline.ContainerStatus
}

func splitByContainers(status *apipipeline.ObjectStatusInfo, pod *v1.Pod, events []*v1.Event) []containerInfo {
	containers := map[string]containerInfo{}

	for _, c := range pod.Spec.Containers {
		containers[c.Name] = containerInfo{
			name: c.Name,
			spec: c,
			platformStatus: &apipipeline.ContainerStatus{
				Type: apipipeline.ContainerType_CONTAINER_TYPE_MAIN,
			},
		}
	}
	for _, c := range pod.Spec.InitContainers {
		ci := containerInfo{
			name:           c.Name,
			spec:           c,
			platformStatus: &apipipeline.ContainerStatus{},
		}
		if r := c.RestartPolicy; r != nil && *r == v1.ContainerRestartPolicyAlways {
			ci.platformStatus.Type = apipipeline.ContainerType_CONTAINER_TYPE_SIDECAR
		} else {
			ci.platformStatus.Type = apipipeline.ContainerType_CONTAINER_TYPE_INIT
		}
		containers[c.Name] = ci
	}

	for _, s := range pod.Status.ContainerStatuses {
		info := containers[s.Name]
		info.status = s
		containers[s.Name] = info
	}
	for _, s := range pod.Status.InitContainerStatuses {
		info := containers[s.Name]
		info.status = s
		containers[s.Name] = info
	}

	for _, e := range events {
		p := e.InvolvedObject.FieldPath
		name := containerNameFromEventFieldPath(p)
		if name == "" {
			continue
		}
		info := containers[name]
		info.events = append(info.events, e)
		containers[name] = info
	}

	for name, c := range containers {
		c.subObj = &apipipeline.SubObjectStatus{
			Name:       name,
			Properties: map[string]string{},
			PlatformStatus: []*apipipeline.PlatformObjectStatus{{
				Name: name,
				Kind: &apipipeline.PlatformObjectStatus_Container{
					Container: c.platformStatus,
				},
			}},
		}
		containers[name] = c
		status.SubObjects = append(status.SubObjects, c.subObj)
	}

	return maps.Values(containers)
}

var _containerNameRe = regexp.MustCompile(`spec\.(initC|c)ontainers{(?P<name>[^}]*)}`)

func containerNameFromEventFieldPath(s string) string {
	m := _containerNameRe.FindStringSubmatch(s)
	idx := _containerNameRe.SubexpIndex("name")
	if idx >= len(m) {
		return ""
	}
	return m[idx]
}

func getCondition(conditions []v1.PodCondition, conditionType string) *v1.PodCondition {
	for _, c := range conditions {
		if c.Type == v1.PodConditionType(conditionType) {
			return &c
		}
	}
	return nil
}

func getEventWithPrefix(events []*v1.Event, eventType, prefix string) *v1.Event {
	for _, e := range events {
		if e.Reason == eventType && strings.HasPrefix(e.Message, prefix) {
			return e
		}
	}
	return nil
}

func getObjectCondition(conditions []*apipipeline.ObjectCondition, conditionName string) *apipipeline.ObjectCondition {
	for _, c := range conditions {
		if c.GetName() == conditionName {
			return c
		}
	}
	return nil
}

func timestampFromCondition(condition *v1.PodCondition) *timestamppb.Timestamp {
	return timestamppb.New(condition.LastTransitionTime.Time)
}

func timestampFromEvent(e *v1.Event) *timestamppb.Timestamp {
	if !e.LastTimestamp.IsZero() {
		return timestamppb.New(e.LastTimestamp.Time)
	}

	if !e.FirstTimestamp.IsZero() {
		return timestamppb.New(e.FirstTimestamp.Time)
	}

	if !e.EventTime.IsZero() {
		return timestamppb.New(e.EventTime.Time)
	}

	return timestamppb.Now()
}

func onDeploymentUpdated(
	obj client.Object,
	_ []*corev1.Event,
	objectWatcher plugin.ObjectWatcher,
) *apipipeline.ObjectStatusInfo {
	dep := obj.(*appsv1.Deployment)
	return OnPodTemplatedUpdated(dep.Spec.Template, objectWatcher)
}

func OnPodTemplatedUpdated(
	template v1.PodTemplateSpec, objectWatcher plugin.ObjectWatcher,
) *apipipeline.ObjectStatusInfo {

	objectWatcher.WatchSecondaryByLabels(PodLabelSelector(template), &corev1.Pod{}, onPodUpdated)
	for _, v := range template.Spec.Volumes {
		if v.ConfigMap != nil {
			objectWatcher.WatchSecondaryByName(v.ConfigMap.Name, &corev1.ConfigMap{}, onConfigMapUpdated)
		} else if v.Secret != nil {
			objectWatcher.WatchSecondaryByName(v.Secret.SecretName, &corev1.Secret{}, onSecretUpdated)
		}
	}

	status := &apipipeline.ObjectStatusInfo{
		Properties: map[string]string{},
	}

	return status
}

func PodLabelSelector(template v1.PodTemplateSpec) labels.Selector {
	selector := labels.NewSelector()
	for key, val := range template.GetLabels() {
		req, err := labels.NewRequirement(key, selection.Equals, []string{val})
		if err != nil {
			// This cannot happen
			panic(err)
		}
		selector = selector.Add(*req)
	}

	req, err := labels.NewRequirement("batch.kubernetes.io/job-name", selection.DoesNotExist, nil)
	if err != nil {
		// This cannot happen
		panic(err)
	}

	return selector.Add(*req)
}

func onConfigMapUpdated(
	obj client.Object,
	_ []*corev1.Event,
	_ plugin.ObjectWatcher,
) *apipipeline.ObjectStatusInfo {
	cm := obj.(*corev1.ConfigMap)

	status := &apipipeline.ObjectStatusInfo{
		Properties: map[string]string{},
		PlatformStatus: []*apipipeline.PlatformObjectStatus{
			{
				Name: cm.GetName(),
				Kind: &apipipeline.PlatformObjectStatus_ConfigFile{
					ConfigFile: &apipipeline.ConfigFileStatus{},
				},
			},
		},
	}

	return status
}

func onSecretUpdated(
	obj client.Object,
	_ []*corev1.Event,
	_ plugin.ObjectWatcher,
) *apipipeline.ObjectStatusInfo {
	secret := obj.(*corev1.Secret)

	status := &apipipeline.ObjectStatusInfo{
		Properties: map[string]string{},
		PlatformStatus: []*apipipeline.PlatformObjectStatus{
			{
				Name: secret.GetName(),
				Kind: &apipipeline.PlatformObjectStatus_ConfigFile{
					ConfigFile: &apipipeline.ConfigFileStatus{},
				},
			},
		},
	}

	return status
}

func onServiceUpdated(
	obj client.Object,
	_ []*corev1.Event,
	_ plugin.ObjectWatcher,
) *apipipeline.ObjectStatusInfo {
	svc := obj.(*corev1.Service)

	var platformStatuses []*apipipeline.PlatformObjectStatus
	for _, p := range svc.Spec.Ports {
		platformStatuses = append(platformStatuses, &apipipeline.PlatformObjectStatus{
			Name: p.Name,
			Kind: &apipipeline.PlatformObjectStatus_Interface{
				Interface: &apipipeline.InterfaceStatus{
					Port: uint32(p.Port),
				},
			},
		})
	}

	return &apipipeline.ObjectStatusInfo{
		Properties:     map[string]string{},
		PlatformStatus: platformStatuses,
	}
}

func (p *Plugin) WatchObjectStatus(ctx context.Context, watcher plugin.CapsuleWatcher) error {
	errChan := make(chan error, 1)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go runWatch(ctx, watcher, &corev1.Service{}, onServiceUpdated, errChan)
	go runWatch(ctx, watcher, &appsv1.Deployment{}, onDeploymentUpdated, errChan)

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return nil
	}
}

func runWatch(ctx context.Context,
	watcher plugin.CapsuleWatcher,
	obj client.Object,
	cb plugin.WatchCallback,
	errChan chan error,
) {
	err := watcher.WatchPrimary(ctx, obj, cb)
	if err != nil {
		errChan <- err
	}
}
