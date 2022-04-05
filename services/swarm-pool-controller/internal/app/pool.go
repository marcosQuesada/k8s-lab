package app

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/k8s-lab/pkg/config"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

const defaultTimeout = time.Second

type Pool interface {
	Size() int
	UpdateSize(context.Context, int) (version int64, err error)
	Dump(ctx context.Context, namespace, configMapName string) error
}

type workloadBalancer interface {
	BalanceWorkload(totalWorkers int, version int64) (*config.Workloads, error)
	Workloads() *config.Workloads
}

// @TODO: Refactor and remove
type delegated interface {
	Assign(ctx context.Context, namespace, configMapName string, w *config.Workloads) error
	RestartWorker(ctx context.Context, namespace, name string) error
}

type pool struct {
	state     workloadBalancer
	delegated delegated
	version   int64
	size      int
	mutex     sync.RWMutex
}

// newWorkerPool instantiates workers pool
func newWorkerPool(version int64, cmp workloadBalancer, not delegated) Pool {
	return &pool{
		version:   version,
		state:     cmp,
		delegated: not,
	}
}

// UpdateSize sets pool expected size
func (p *pool) UpdateSize(ctx context.Context, newSize int) (version int64, err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.size == newSize {
		return p.version, nil
	}

	previousSize := p.size
	p.size = newSize
	p.version++

	log.Infof("Pool Version Update %d Size From %d to %d", p.version, previousSize, newSize)

	_, err = p.state.BalanceWorkload(newSize, p.version)
	if err != nil {
		return p.version, fmt.Errorf("err on balance workload %v", err)
	}

	return p.version, nil
}

func (p *pool) Dump(ctx context.Context, namespace, configMapName string) error {
	wkl := p.state.Workloads()
	if err := p.delegated.Assign(ctx, namespace, configMapName, wkl); err != nil {
		return fmt.Errorf("unable to dump workload on namespace %s configMapName %s error %v", namespace, configMapName, err)
	}

	return nil
}

func (p *pool) Size() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.size
}
