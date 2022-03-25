package v1alpha1

import (
	"github.com/marcosQuesada/k8s-lab/pkg/operator/crd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CrdKind   string = "ConfigMapClaimOwner"
	Version   string = "v1alpha1"
	Singular  string = "configmapownerclaim"
	Plural    string = "configmapownerclaims"
	ShortName string = "cmoc"
	Name      string = Plural + "." + crd.GroupName
)

const (
	Deployment  = "Deployment" // @TODO: Refactor, unify
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
