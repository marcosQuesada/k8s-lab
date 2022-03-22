package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

const (
	Deployment  = "Deployment"
	StatefulSet = "StatefulSet"
)

// PoolStatus defines pool subject state
type PoolStatus struct {
	Phase string `json:"phase,omitempty"`
}

// ConfigMapPodsRefresherSpec defines the desired state of Swarm
type ConfigMapPodsRefresherSpec struct {
	Version          int64  `json:"version"`
	Namespace        string `json:"namespace"`
	WatchedConfigMap string `json:"watched-config-map"`
	PoolType         string `json:"pool-type"`
	PoolSubjectName  string `json:"pool-subject-name"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConfigMapPodRefresher defines confimap and pool subject reflecting status
type ConfigMapPodRefresher struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigMapPodsRefresherSpec `json:"spec,omitempty"`
	Status PoolStatus                 `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConfigMapPodRefresherList contains a list of ConfigMapPodRefresher
type ConfigMapPodRefresherList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConfigMapPodRefresher `json:"items"`
}
