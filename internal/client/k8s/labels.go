package k8s

import "github.com/rigdev/rig/internal/gateway/cluster"

const (
	labelName         = "app.kubernetes.io/name"
	labelInstance     = "app.kubernetes.io/instance"
	labelManagedBy    = "app.kubernetes.io/managed-by"
	labelManagedByRig = "rig"
	labelRigCapsuleID = "rig.dev/capsule-id"
)

func selectorLabels(capsuleID string) map[string]string {
	return map[string]string{
		labelName:     capsuleID,
		labelInstance: capsuleID,
	}
}

func commonLabels(capsuleID string, c *cluster.Capsule) map[string]string {
	ls := selectorLabels(capsuleID)
	ls[labelManagedBy] = labelManagedByRig
	ls[labelRigCapsuleID] = c.CapsuleID
	return ls
}

func capsuleIDFromLabels(labels map[string]string) string {
	return labels[labelRigCapsuleID]
}
