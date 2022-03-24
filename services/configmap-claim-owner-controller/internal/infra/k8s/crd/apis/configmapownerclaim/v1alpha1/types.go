package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

const (
	Deployment  = "Deployment" // @TODO: Avoid duplication
	StatefulSet = "StatefulSet"
)

// ConfigMapClaimOwnerSpec defines the desired state of Swarm
type ConfigMapClaimOwnerSpec struct {
	Namespace string `json:"namespace"`
	ConfigMap string `json:"config-map"`
	OwnerType string `json:"owner-type"`
	OwnerName string `json:"owner-name"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConfigMapClaimOwner defines claimed confimaps
type ConfigMapClaimOwner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ConfigMapClaimOwnerSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConfigMapClaimOwnerList contains a list of ConfigMapClaimOwner
type ConfigMapClaimOwnerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConfigMapClaimOwner `json:"items"`
}
