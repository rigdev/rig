//nolint:revive
package ingress_routes

import (
	"context"
	"fmt"
	"strings"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	v1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	apipipeline "github.com/rigdev/rig-go-api/operator/api/v1/pipeline"
	"github.com/rigdev/rig/pkg/controller/plugin"
	"github.com/rigdev/rig/pkg/pipeline"
	"google.golang.org/protobuf/types/known/timestamppb"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func toIngressStatus(ingress *netv1.Ingress) *apipipeline.ObjectStatusInfo {
	status := &apipipeline.ObjectStatusInfo{
		Properties: map[string]string{},
	}

	var hosts []string
	host := ""
	for _, r := range ingress.Spec.Rules {
		if r.Host != "" {
			hosts = append(hosts, r.Host)
			host = r.Host
		}
	}

	status.Properties["Hosts"] = strings.Join(hosts, ", ")

	ipCondition := &apipipeline.ObjectCondition{
		Name:    "LoadBalancer IP",
		State:   apipipeline.ObjectState_OBJECT_STATE_PENDING,
		Message: "Waiting for IP assignment",
	}

	for _, lb := range ingress.Status.LoadBalancer.Ingress {
		if lb.IP != "" {
			ipCondition.State = apipipeline.ObjectState_OBJECT_STATE_HEALTHY
			ipCondition.Message = fmt.Sprintf("IP Assigned '%s'", lb.IP)
			status.Properties["IP"] = lb.IP
		}
	}

	parts := strings.Split(ingress.GetName(), "-")
	routeID := parts[len(parts)-1]
	status.Conditions = append(status.Conditions, ipCondition)
	status.PlatformStatus = append(status.PlatformStatus, &apipipeline.PlatformObjectStatus{
		Name: routeID,
		Kind: &apipipeline.PlatformObjectStatus_Route{
			Route: &apipipeline.RouteStatus{
				Id:            routeID,
				Host:          host,
				InterfaceName: ingress.GetLabels()[pipeline.RigDevInterfaceLabel],
			},
		},
	})
	return status
}

func onCertificateUpdated(
	obj client.Object,
	events []*corev1.Event,
	objectWatcher plugin.ObjectWatcher,
) *apipipeline.ObjectStatusInfo {
	cert := obj.(*cmv1.Certificate)

	revision := 1
	if cert.Status.Revision != nil {
		revision = *cert.Status.Revision
	}
	requestName := fmt.Sprintf("%s-%d", cert.GetName(), revision)
	objectWatcher.WatchSecondaryByName(requestName, &cmv1.CertificateRequest{}, onCertificateRequestUpdated)

	status := &apipipeline.ObjectStatusInfo{
		Properties: map[string]string{},
	}

	for _, c := range cert.Status.Conditions {
		cond := &apipipeline.ObjectCondition{
			Name:      "Certificate " + strings.ToLower(string(c.Type)),
			State:     apipipeline.ObjectState_OBJECT_STATE_PENDING,
			Message:   c.Message,
			UpdatedAt: timestamppb.New(c.LastTransitionTime.Time),
		}

		if c.Status == v1.ConditionTrue {
			cond.State = apipipeline.ObjectState_OBJECT_STATE_HEALTHY
		}

		switch c.Type {
		case "Issuing":
			cond.Name = "Certificate issuing"
		case "Ready":
			cond.Name = "Certificate readying"
		}

		status.Conditions = append(status.Conditions, cond)
	}

	return status
}

func onCertificateRequestUpdated(
	obj client.Object,
	events []*corev1.Event,
	objectWatcher plugin.ObjectWatcher,
) *apipipeline.ObjectStatusInfo {
	certReq := obj.(*cmv1.CertificateRequest)

	status := &apipipeline.ObjectStatusInfo{
		Properties: map[string]string{},
	}

	for _, c := range certReq.Status.Conditions {
		cond := &apipipeline.ObjectCondition{
			Name:      "Certificate request " + strings.ToLower(string(c.Type)),
			State:     apipipeline.ObjectState_OBJECT_STATE_PENDING,
			Message:   c.Message,
			UpdatedAt: timestamppb.New(c.LastTransitionTime.Time),
		}

		if c.Status == v1.ConditionTrue {
			cond.State = apipipeline.ObjectState_OBJECT_STATE_HEALTHY
		}

		switch c.Type {
		case "Approved":
			cond.Name = "Certificate request approval"
		case "Ready":
			cond.Name = "Certificate request readying"
		}

		status.Conditions = append(status.Conditions, cond)
	}

	return status
}

func onIngressUpdated(
	obj client.Object,
	events []*corev1.Event,
	objectWatcher plugin.ObjectWatcher,
) *apipipeline.ObjectStatusInfo {
	ingress := obj.(*netv1.Ingress)

	objectWatcher.WatchSecondaryByLabels(labels.Set{
		pipeline.LabelOwnedByCapsule: ingress.GetLabels()[pipeline.LabelOwnedByCapsule],
	}, &cmv1.Certificate{}, onCertificateUpdated)

	return toIngressStatus(ingress)
}

func (p *Plugin) WatchObjectStatus(ctx context.Context, watcher plugin.CapsuleWatcher) error {
	return watcher.WatchPrimary(ctx, &netv1.Ingress{}, onIngressUpdated)
}
