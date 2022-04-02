package k8s

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/k8s-lab/pkg/config"
	"github.com/marcosQuesada/k8s-lab/pkg/operator"
	"github.com/marcosQuesada/k8s-lab/pkg/operator/pod"
	"github.com/marcosQuesada/k8s-lab/pkg/operator/statefulset"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/app"
	pod2 "github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/pod"
	statefulset2 "github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/statefulset"
	"k8s.io/client-go/kubernetes"
	"net"
	"sync"
)

type Controller interface {
	UpdateSize(newSize int)
	AddWorkerIfNotExists(idx int, name string, IP net.IP) bool
	RemoveWorkerByName(name string)
	Terminate()
}

type delegated interface {
	Assign(ctx context.Context, w *config.Workloads) error
	RestartWorker(ctx context.Context, namespace, name string) error
}

type controllerFactory struct {
	client    kubernetes.Interface
	delegated delegated
	index     map[string]chan struct{}
	mutex     sync.RWMutex
}

func NewWorkerPoolFactory(client kubernetes.Interface, d delegated) *controllerFactory {
	return &controllerFactory{
		client:    client,
		delegated: d,
		index:     map[string]chan struct{}{},
	}
}

// @TODO: Segregate controller builder to allow proper testing...
func (f *controllerFactory) BootController(namespace, watchedLabel string, jobs []config.Job) app.Pool {
	st := app.NewState(jobs, watchedLabel)
	a := app.NewWorkerPool(st, f.delegated, namespace)

	podLwa := pod.NewListWatcherAdapter(f.client, namespace)
	podH := pod2.NewHandler(a)
	podCtl := operator.Build(podLwa, podH, watchedLabel)

	stsLwa := statefulset.NewListWatcherAdapter(f.client, namespace)

	stsH := statefulset2.NewHandler(a, nil) //@TODO: REMOVE
	stsCtl := operator.Build(stsLwa, stsH, watchedLabel)

	stopCh := make(chan struct{})
	go podCtl.Run(stopCh)
	go stsCtl.Run(stopCh)

	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.index[key(namespace, watchedLabel)] = stopCh

	return a
}

func (f *controllerFactory) Terminate(namespace, watchedLabel string) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	f.terminate(namespace, watchedLabel)
}

func (f *controllerFactory) Shutdown() {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	for k, c := range f.index {
		close(c)
		delete(f.index, k)
	}
}

func (f *controllerFactory) terminate(namespace, watchedLabel string) {
	k := key(namespace, watchedLabel)
	v, ok := f.index[k]
	if !ok {
		return
	}

	close(v)
	delete(f.index, k)
}

// @TODO: SEGREGATE IT!
func key(namespace, label string) string {
	return fmt.Sprintf("%s/%s", namespace, label)
}
