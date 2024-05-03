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
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
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
			ipCondition.Message = "IP Assigned"
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
				Id:   routeID,
				Host: host,
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

	status := &apipipeline.ObjectStatusInfo{
		Properties: map[string]string{},
	}

	for _, c := range cert.Status.Conditions {
		switch c.Type {
		case "Issuing":
			cond := &apipipeline.ObjectCondition{
				Name:    "Certificate issuing",
				State:   apipipeline.ObjectState_OBJECT_STATE_PENDING,
				Message: c.Message,
			}

			if c.Status == v1.ConditionTrue {
				cond.State = apipipeline.ObjectState_OBJECT_STATE_HEALTHY
			}

			status.Conditions = append(status.Conditions, cond)
		case "Ready":
			cond := &apipipeline.ObjectCondition{
				Name:    "Certificate ready",
				State:   apipipeline.ObjectState_OBJECT_STATE_PENDING,
				Message: c.Message,
			}

			if c.Status == v1.ConditionTrue {
				cond.State = apipipeline.ObjectState_OBJECT_STATE_HEALTHY
			}

			status.Conditions = append(status.Conditions, cond)
		}
	}

	return status
}

func onIngressUpdated(
	obj client.Object,
	events []*corev1.Event,
	objectWatcher plugin.ObjectWatcher,
) *apipipeline.ObjectStatusInfo {
	ingress := obj.(*netv1.Ingress)

	objectWatcher.WatchSecondaryByName(ingress.GetName(), &cmv1.Certificate{}, onCertificateUpdated)
	objectWatcher.WatchSecondaryByName(fmt.Sprint(ingress.GetName(), "-tls"), &cmv1.Certificate{}, onCertificateUpdated)

	return toIngressStatus(ingress)
}

func (p *Plugin) WatchObjectStatus(ctx context.Context, watcher plugin.CapsuleWatcher) error {
	return watcher.WatchPrimary(ctx, &netv1.Ingress{}, onIngressUpdated)
}
