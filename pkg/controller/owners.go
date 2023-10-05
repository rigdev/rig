package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func IsOwnedBy(owner metav1.Object, obj metav1.Object) bool {
	for _, ref := range obj.GetOwnerReferences() {
		if ref.UID == owner.GetUID() &&
			ref.Controller != nil &&
			*ref.Controller {
			return true
		}
	}
	return false
}
