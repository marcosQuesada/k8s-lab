package v1alpha1

import (
	"github.com/marcosQuesada/k8s-lab/pkg/operator/crd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CrdKind   string = "Swarm"
	Version   string = "v1alpha1"
	Singular  string = "swarm"
	Plural    string = "swarms"
	ShortName string = "swm"
	Name      string = Plural + "." + crd.GroupName
)

const (
	PhasePending = "PENDING"
	PhaseRunning = "RUNNING"
	PhaseDone    = "DONE"
)

type Job string

// Status defines the observed state of Worker
type Status struct {
	Phase string `json:"phase,omitempty"`
}

type Worker struct {
	Name      string `json:"name"`
	Jobs      []Job  `json:"jobs"`
	CreatedAt int64  `json:"created_at"`
	State     Status `json:"state"`
}

// SwarmSpec defines the desired state of Swarm
type SwarmSpec struct {
	Namespace       string   `json:"namespace"`
	StatefulSetName string   `json:"statefulset-name"`
	ConfigMapName   string   `json:"configmap-name"`
	Version         int64    `json:"version"`
	Workload        []Job    `json:"workload"`
	Size            int      `json:"size, omitempty"`
	Members         []Worker `json:"members,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Swarm runs a command Swarm a given schedule.
type Swarm struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SwarmSpec `json:"spec,omitempty"`
	Status Status    `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SwarmList contains a list of Swarm
type SwarmList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Swarm `json:"items"`
}
