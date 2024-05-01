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
	"google.golang.org/protobuf/types/known/timestamppb"
	netv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func toIngressStatus(ingress *netv1.Ingress) *apipipeline.ObjectStatus {
	status := &apipipeline.ObjectStatus{
		Type:       apipipeline.ObjectType_OBJECT_TYPE_PRIMARY,
		State:      apipipeline.ObjectState_OBJECT_STATE_HEALTHY,
		UpdatedAt:  timestamppb.Now(),
		Properties: map[string]string{},
	}

	var hosts []string
	for _, r := range ingress.Spec.Rules {
		if r.Host != "" {
			hosts = append(hosts, r.Host)
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
	status.Conditions = append(status.Conditions, ipCondition)

	return status
}

func onCertificateUpdated(obj client.Object, objectWatcher plugin.ObjectWatcher) *apipipeline.ObjectStatus {
	cert := obj.(*cmv1.Certificate)

	status := &apipipeline.ObjectStatus{
		Type:       apipipeline.ObjectType_OBJECT_TYPE_SECONDARY,
		State:      apipipeline.ObjectState_OBJECT_STATE_HEALTHY,
		UpdatedAt:  timestamppb.Now(),
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

func onIngressUpdated(obj client.Object, objectWatcher plugin.ObjectWatcher) *apipipeline.ObjectStatus {
	ingress := obj.(*netv1.Ingress)

	objectWatcher.WatchSecondaryByName(ingress.GetName(), &cmv1.Certificate{}, onCertificateUpdated)
	objectWatcher.WatchSecondaryByName(fmt.Sprint(ingress.GetName(), "-tls"), &cmv1.Certificate{}, onCertificateUpdated)

	return toIngressStatus(ingress)
}

func (p *Plugin) WatchObjectStatus(ctx context.Context, watcher plugin.CapsuleWatcher) error {
	return watcher.WatchPrimary(ctx, &netv1.Ingress{}, onIngressUpdated)
}
