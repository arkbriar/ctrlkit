package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CronJobSpec struct {
	Suspend *bool `json:"suspend,omitempty"`
}

type CronJobStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type CronJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CronJobSpec   `json:"spec"`
	Status            CronJobStatus `json:"status"`
}

// +kubebuilder:object:root=true

type CronJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CronJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CronJob{}, &CronJobList{})
}
