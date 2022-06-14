// +kubebuilder:object:generate=true
// +groupName=demo

// Package v1 containers API schema definitions for the demo v1 API group.
package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register objects.
	GroupVersion = schema.GroupVersion{
		Group:   "demo",
		Version: "v1",
	}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
