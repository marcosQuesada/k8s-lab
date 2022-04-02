package app

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/k8s-lab/pkg/config"
	log "github.com/sirupsen/logrus"
	"sort"
	"sync"
	"time"
)

const defaultTimeout = time.Second

type Pool interface {
	Size() int
	UpdateSize(newSize int)
	AddWorkerIfNotExists(idx int, namespace, name string) bool
	RemoveWorkerByName(namespace, name string)
}

type workloadBalancer interface {
	BalanceWorkload(totalWorkers int, version int64) error
	Workloads() *config.Workloads
}

// @TODO: Refactor and remove
type delegated interface {
	Assign(ctx context.Context, w *config.Workloads) error
	RestartWorker(ctx context.Context, namespace, name string) error
}

type pool struct {
	index     map[string]*worker // @TODO: Move index to Key[namespace/name]
	state     workloadBalancer
	delegated delegated
	version   int64
	size      int
	mutex     sync.RWMutex
}

// NewWorkerPool instantiates workers pool
func NewWorkerPool(version int64, cmp workloadBalancer, not delegated) Pool {
	return &pool{
		index:     make(map[string]*worker),
		version:   version,
		state:     cmp,
		delegated: not,
	}
}

func (p *pool) Size() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.size
}

// UpdateSize sets pool expected size
func (p *pool) UpdateSize(newSize int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.size == newSize {
		return
	}

	previousSize := p.size
	p.size = newSize
	p.version++

	log.Infof("Pool Version Update %d Size From %d to %d", p.version, previousSize, newSize)

	if err := p.state.BalanceWorkload(newSize, p.version); err != nil {
		log.Errorf("err on balance started %v", err)
	}

	if previousSize == 0 {
		return
	}

	ws := p.geAllWorkers()
	totalToRefresh := newSize
	if newSize > previousSize {
		totalToRefresh = previousSize
	}

	if len(ws) < totalToRefresh {
		totalToRefresh = len(ws)
	}

	for i := 0; i < totalToRefresh; i++ {
		w := ws[i]
		w.MarkToRefresh()
	}

	log.Infof("Total %d Workers marked to Refresh %d to version %d", len(ws), totalToRefresh, p.version)
}

// AddWorkerIfNotExists register a worker if not exists in the pool
func (p *pool) AddWorkerIfNotExists(idx int, namespace, name string) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, ok := p.index[name]; ok {
		log.Infof("pod %s already on pool", name)
		return false
	}

	p.index[name] = newWorker(idx, namespace, name, p.delegated)

	log.Debugf("Added worker to Pool name %s length %d, size %d", name, len(p.index), p.size)

	return true
}

// RemoveWorkerByName removes worker from pool
func (p *pool) RemoveWorkerByName(namespace, name string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	log.Infof("Removing worker %s from Pool name", name)
	delete(p.index, name)
}

func (p *pool) requestRestart(ctx context.Context, w *worker) error {
	log.Infof("Scheduling worker %s refresh", w.name)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := w.delegated.RestartWorker(ctx, w.namespace, w.name); err != nil {
		log.Errorf("unable to restart worker, error %v", err)
		return err
	}
	w.MarkRefreshed()
	return nil
}

func (p *pool) geAllWorkers() []*worker {
	var res workerList
	for _, w := range p.index {
		res = append(res, w)
	}
	sort.Sort(res)

	return res
}

func (p *pool) worker(name string) (*worker, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	v, ok := p.index[name]
	if !ok {
		log.Infof("pod %s already on pool", name)
		return nil, fmt.Errorf("no worker %s found", name)
	}
	return v, nil
}
