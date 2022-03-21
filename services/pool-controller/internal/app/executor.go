package app

import (
	"context"
	"fmt"
	ap "github.com/marcosQuesada/k8s-lab/pkg/config"
	log "github.com/sirupsen/logrus"
	"net"
)

type workerProvider interface {
	Assignation(ctx context.Context, IP net.IP) (*ap.Workload, error)
}

type workerManager interface {
	Refresh(ctx context.Context, podName string) error
}

type delegatedStorage interface {
	Set(ctx context.Context, a *ap.Workloads) error
}

type executor struct {
	storage delegatedStorage
	remotes workerProvider
	manager workerManager
}

func NewExecutor(s delegatedStorage, p workerProvider, m workerManager) *executor {
	return &executor{
		storage: s,
		remotes: p,
		manager: m,
	}
}

func (e *executor) Assign(ctx context.Context, w *ap.Workloads) (err error) {
	log.Infof("Persist Workload version %d to assign to %v", w.Version, w.Workloads)
	return e.storage.Set(ctx, w)
}

// @TODO: refactor and remove
func (e *executor) Assignation(ctx context.Context, w *worker) (a *ap.Workload, err error) {
	log.Infof("Get config to %s IP %s", w.name, w.IP.String())
	res, err := e.remotes.Assignation(ctx, w.IP)
	if err != nil {
		return nil, fmt.Errorf("unable to get remote assignation on %s error %v", w.name, err)
	}
	log.Infof("Received config from %s IP %s version %d jobs %v", w.name, w.IP.String(), res.Version, len(res.Jobs))
	return res, nil
}

func (e *executor) RestartWorker(ctx context.Context, name string) error {
	log.Infof("Restarting worker %s", name)
	return e.manager.Refresh(ctx, name)
}
