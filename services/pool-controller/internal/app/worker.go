package app

import (
	log "github.com/sirupsen/logrus"
	"net"
	"sync"
)

type worker struct {
	name       string
	IP         net.IP
	index      int
	version    int64
	state      State
	stateMutex sync.RWMutex
	delegated  delegated
}

func newWorker(idx int, name string, IP net.IP, d delegated) *worker {
	return &worker{
		name:      name,
		IP:        IP,
		index:     idx,
		state:     WaitingAssignation,
		delegated: d,
	}
}

// NeedsRefresh state refresh checker
func (w *worker) NeedsRefresh() bool {
	w.stateMutex.RLock()
	defer w.stateMutex.RUnlock()

	return w.state == NeedsRefresh
}

// MarkToRefresh sets worker in waiting refresh mode
func (w *worker) MarkToRefresh() {
	log.Infof("worker %s marked to refresh", w.name)
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()
	w.state = NeedsRefresh
}

// MarkRefreshed updates state to Syncing
func (w *worker) MarkRefreshed() {
	log.Infof("worker %s marked refreshed", w.name)
	w.stateMutex.Lock()
	defer w.stateMutex.Unlock()
	w.state = Syncing
}

type workerList []*worker

// Len method to get an ordered worker list
func (e workerList) Len() int {
	return len(e)
}

// Less method to get an ordered worker list
func (e workerList) Less(i, j int) bool {
	return e[i].index < e[j].index
}

// Swap method to get an ordered worker list
func (e workerList) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
