package app

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/k8s-lab/pkg/config"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/apis/swarm/v1alpha1"
	v1alpha1Lister "github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/generated/listers/swarm/v1alpha1"
	log "github.com/sirupsen/logrus"
	"sync"
)

type manager struct {
	index       map[string]Pool
	mutex       sync.RWMutex
	delegated   delegated
	swarmLister v1alpha1Lister.SwarmLister
}

func NewManager(d delegated, l v1alpha1Lister.SwarmLister) *manager {
	return &manager{
		index:       make(map[string]Pool),
		delegated:   d,
		swarmLister: l,
	}
}

func (m *manager) Process(ctx context.Context, namespace, name string, version int64, workloads []v1alpha1.Job) {
	log.Infof("Adding swarm namespace %s name %s version %d total workloads %d", namespace, name, version, len(workloads))
	m.mutex.Lock()
	defer m.mutex.Unlock()

	k := namespace + "/" + name
	wp := []config.Job{}
	for _, w := range workloads {
		wp = append(wp, config.Job(w))
	}
	ast := newState(wp, k)
	m.index[k] = newWorkerPool(version, ast, m.delegated)
}

func (m *manager) UpdateSize(ctx context.Context, namespace, name string, size int) (int64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	k := namespace + "/" + name
	if _, ok := m.index[k]; !ok {
		return 0, fmt.Errorf("no %s %s regustered", namespace, name)
	}

	v, err := m.index[k].UpdateSize(ctx, size)
	if err != nil {
		return v, err
	}

	sw, err := m.swarmLister.Swarms(namespace).Get(name)
	if err != nil {
		return 0, fmt.Errorf("unable to find swarm %s error %v", name, err)
	}

	if err := m.index[k].Dump(ctx, namespace, sw.Spec.ConfigMapName); err != nil {
		return v, fmt.Errorf("unable to dump swarm %s error %v", name, err)
	}

	return v, nil
}

func (m *manager) Delete(ctx context.Context, namespace, name string) {
	log.Infof("Delete swarm namespace %s name %s", namespace, name)
	m.mutex.Lock()
	defer m.mutex.Unlock()

	k := namespace + "/" + name
	if _, ok := m.index[k]; !ok {
		return
	}

	delete(m.index, k)
}
