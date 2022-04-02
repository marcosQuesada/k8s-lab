package app

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/k8s-lab/pkg/config"
	log "github.com/sirupsen/logrus"
	"net"
	"sort"
	"sync"
	"time"
)

const defaultSleepBeforeNotify = time.Second * 2
const defaultTimeout = time.Second * 1
const DefaultWorkerFrequency = time.Second * 2

type Pool interface {
	Size() int
	UpdateSize(newSize int)
	AddWorkerIfNotExists(idx int, name string, IP net.IP) bool
	RemoveWorkerByName(name string)
	Terminate()
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
	index          map[string]*worker
	namespace      string
	state          workloadBalancer
	delegated      delegated
	version        int64
	expectedSize   int
	underVariation bool
	stopChan       chan struct{}
	mutex          sync.RWMutex
}

// NewWorkerPool instantiates workers pool
func NewWorkerPool(namespace string, version int64, cmp workloadBalancer, not delegated) Pool {
	s := &pool{
		index:          make(map[string]*worker),
		namespace:      namespace,
		version:        version,
		state:          cmp,
		delegated:      not,
		underVariation: true,
		stopChan:       make(chan struct{}),
	}

	//go wait.Until(s.conciliate, DefaultWorkerFrequency, s.stopChan)

	return s
}

func (p *pool) Size() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.expectedSize
}

// UpdateSize sets pool expected size
func (p *pool) UpdateSize(newSize int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.expectedSize == newSize {
		return
	}

	previousSize := p.expectedSize
	p.expectedSize = newSize
	p.underVariation = true
	p.version++

	log.Infof("Pool Version Update %d Size From %d to %d", p.version, previousSize, newSize)

	if err := p.state.BalanceWorkload(newSize, p.version); err != nil {
		log.Errorf("err on balance started %v", err)
	}

	//if err := p.delegated.Assign(context.Background(), p.state.Workloads()); err != nil {
	//	log.Errorf("config error %v", err)
	//} //@TODO:

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
func (p *pool) AddWorkerIfNotExists(idx int, name string, IP net.IP) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if _, ok := p.index[name]; ok {
		log.Infof("pod %s already on pool", name)
		return false
	}

	p.index[name] = newWorker(idx, name, p.delegated)

	log.Debugf("Added worker to Pool name %s IP %s length %d, expectedSize %d", name, IP, len(p.index), p.expectedSize)

	return true
}

// RemoveWorkerByName removes worker from pool
func (p *pool) RemoveWorkerByName(name string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	log.Infof("Removing worker %s from Pool name", name)
	delete(p.index, name)
}

// Terminate stops conciliation loop
func (p *pool) Terminate() {
	close(p.stopChan)
}

func (p *pool) conciliate() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.expectedSize == 0 {
		return
	}

	if !p.underVariation {
		return
	}

	log.Infof("state contiliation version %d, expected slaves: %d on index %d", p.version, p.expectedSize, len(p.index))

	for _, w := range p.geAllWorkers() {
		if !w.NeedsRefresh() {
			continue
		}

		log.Infof("Request scheduled restart to %s", w.name)
		go func() {
			if err := p.requestRestart(context.Background(), w); err != nil {
				log.Errorf("unable to restart worker %s error %v", w.name, err)
			}
		}()
	}

	log.Info("Stopping conciliation loop, variation completed!")
	p.underVariation = false
}

func (p *pool) requestRestart(ctx context.Context, w *worker) error {
	log.Infof("Scheduling worker %s refresh", w.name)
	//time.Sleep(defaultSleepBeforeNotify * time.Duration(2*(1+w.index))) // @TODO: Address it!
	//
	//ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	//defer cancel()
	//if err := w.delegated.RestartWorker(ctx, p.namespace, w.name); err != nil {
	//	log.Errorf("unable to restart worker, error %v", err)
	//	return err
	//}
	//w.MarkRefreshed()
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
