package app

import (
	"fmt"
	"github.com/marcosQuesada/k8s-lab/services/config-reloader-controller/internal/infra/k8s/crd/apis/configmappodrefresher/v1alpha1"
	"sync"
)

type PoolType string

const Deployment = PoolType("deployment")
const StatefulSet = PoolType("statefulset")

type entry struct {
	Namespace     string
	ConfigMapName string
	PoolType      PoolType
	PoolName      string
}

func newEntry(namespace, name string, pt PoolType, pm string) *entry {
	return &entry{
		Namespace:     namespace,
		ConfigMapName: name,
		PoolType:      pt,
		PoolName:      pm,
	}
}

func (e *entry) String() string {
	return fmt.Sprintf("%s/%s", e.Namespace, e.ConfigMapName)
}

type App struct {
	index map[string]struct{}
	mutex sync.RWMutex
}

func NewApp() *App {
	return &App{
		index: map[string]struct{}{},
	}
}

func (a *App) Watch(c *v1alpha1.ConfigMapPodRefresher) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	e := newEntry(
		c.Spec.Namespace,
		c.Spec.WatchedConfigMap,
		PoolType(c.Spec.PoolType),
		c.Spec.PoolSubjectName)

	if _, ok := a.index[e.String()]; ok {
		return
	}

	a.index[e.String()] = struct{}{}
}

func (a *App) IsRegistered(namespace, name string) bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	_, ok := a.index[newEntry(namespace, name, "", "").String()]

	return ok
}

func (a *App) key(namespace, name string) string {
	return ""
}
