package ctrlkit

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ValidateOwnership(obj, owner client.Object) bool {
	ref := metav1.GetControllerOfNoCopy(obj)
	if ref == nil {
		return false
	}
	if ref.UID == owner.GetUID() {
		return true
	}
	return false
}
