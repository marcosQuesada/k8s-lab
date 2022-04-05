package app

import (
	"context"
	ap "github.com/marcosQuesada/k8s-lab/pkg/config"
	log "github.com/sirupsen/logrus"
)

type workerManager interface {
	Refresh(ctx context.Context, namespace, podName string) error
}

type delegatedStorage interface {
	Set(ctx context.Context, namespace, configMapName string, a *ap.Workloads) error
}

type executor struct {
	storage delegatedStorage
	manager workerManager
}

func NewExecutor(s delegatedStorage, m workerManager) *executor {
	return &executor{
		storage: s,
		manager: m,
	}
}

func (e *executor) Assign(ctx context.Context, namespace, configMapName string, w *ap.Workloads) (err error) {
	log.Infof("Persist Workload version %d to assign to %v", w.Version, w.Workloads)
	return e.storage.Set(ctx, namespace, configMapName, w)
}

func (e *executor) RestartWorker(ctx context.Context, namespace, name string) error {
	log.Infof("Restarting worker %s", name)
	return e.manager.Refresh(ctx, namespace, name)
}

type nopExecutor struct {
}

func NewNopExecutor() *nopExecutor {
	return &nopExecutor{}
}

func (e *nopExecutor) Assign(ctx context.Context, w *ap.Workloads) error {
	log.Infof("Persist Workload version %d to assign to %v", w.Version, w.Workloads)
	return nil
}

func (e *nopExecutor) RestartWorker(ctx context.Context, namespace, name string) error {
	log.Infof("Restarting worker %s", name)
	return nil
}
