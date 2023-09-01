package k8s

import "github.com/rigdev/rig/internal/gateway/cluster"

const (
	labelName            = "app.kubernetes.io/name"
	labelInstance        = "app.kubernetes.io/instance"
	labelVersion         = "app.kubernetes.io/version"
	labelManagedBy       = "app.kubernetes.io/managed-by"
	labelManagedByRig = "rig"
	labelRigCapsuleID = "rig.dev/capsule-id"
)

func selectorLabels(capsuleName string) map[string]string {
	return map[string]string{
		labelName:     capsuleName,
		labelInstance: capsuleName,
	}
}

func commonLabels(capsuleName string, c *cluster.Capsule) map[string]string {
	ls := selectorLabels(capsuleName)
	ls[labelVersion] = c.BuildID
	ls[labelManagedBy] = labelManagedByRig
	ls[labelRigCapsuleID] = c.CapsuleID
	return ls
}

func capsuleIDFromLabels(labels map[string]string) string {
	return labels[labelRigCapsuleID]
}
