package app

import (
	"fmt"
	"github.com/marcosQuesada/k8s-lab/pkg/config"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/apis/swarm/v1alpha1"
	log "github.com/sirupsen/logrus"
	"sync"
)

type manager struct {
	index     map[string]Pool
	mutex     sync.RWMutex
	delegated delegated
}

func NewManager(d delegated) *manager {
	return &manager{
		index:     make(map[string]Pool),
		delegated: d,
	}
}

func (m *manager) Add(namespace, label string, version int64, workloads []v1alpha1.Job) {
	log.Infof("Adding swarm namespace %s label %s version %d workloads %v", namespace, label, version, workloads)
	m.add(namespace, label, version, workloads)
}

func (m *manager) Update(namespace, label string, version int64, workloads []v1alpha1.Job) {
	log.Infof("Update swarm namespace %s label %s version %d workloads %v", namespace, label, version, workloads)
}

func (m *manager) Delete(namespace, label string) {
	log.Infof("Delete swarm namespace %s label %s", namespace, label)
	m.mutex.Lock()
	defer m.mutex.Unlock()
	k := key(namespace, label)
	if _, ok := m.index[k]; !ok {
		return
	}

	delete(m.index, k)
}

// @TODO: version!
func (m *manager) add(namespace, label string, version int64, workloads []v1alpha1.Job) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	k := key(namespace, label)
	if _, ok := m.index[k]; ok {
		return
	}

	wp := []config.Job{}
	for _, w := range workloads {
		wp = append(wp, config.Job(w))
	}

	log.Infof("Booting controller on namespace %s label %s total workloads %d", namespace, label, len(wp))
	ast := NewState(wp, label)
	m.index[k] = NewWorkerPool(namespace, version, ast, m.delegated)
}

// @TODO: Unify!
func key(namespace, label string) string {
	return fmt.Sprintf("%s/%s", namespace, label)
}
