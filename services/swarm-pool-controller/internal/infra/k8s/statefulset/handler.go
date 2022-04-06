package statefulset

import (
	"context"
	log "github.com/sirupsen/logrus"
	api "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sync"
)

// PoolController models an ordered set of workers
type PoolController interface {
	UpdatePoolSize(ctx context.Context, namespace, name string, size int) error
}

// Handler handles statefulset state updates
type Handler struct {
	controller    PoolController
	selector      SelectorStore
	lastSizeIndex map[string]int
	mutex         sync.RWMutex
}

// NewHandler instantiates statefulset handler
func NewHandler(c PoolController, s SelectorStore) *Handler {
	return &Handler{
		controller:    c,
		selector:      s,
		lastSizeIndex: map[string]int{},
	}
}

func (h *Handler) Handle(ctx context.Context, o runtime.Object) error {
	ss := o.(*api.StatefulSet)
	if !h.selector.Matches(ss.Namespace, ss.Name, ss.Labels) {
		log.Infof("Skipped sts event namespace %s name %s labels %v", ss.Namespace, ss.Name, ss.Labels)
		return nil
	}

	if !h.hasLastSizeVariation(ss) {
		return nil
	}

	log.Infof("Handle Statefulset Namespace %s name %s Replicas %d", ss.Namespace, ss.Name, uint64(*ss.Spec.Replicas))

	return h.controller.UpdatePoolSize(ctx, ss.Namespace, ss.Name, int(*ss.Spec.Replicas))
}

func (h *Handler) Delete(ctx context.Context, namespace, name string) error {
	log.Infof("Deleted StatefulSet Namespace %s name %s", namespace, name)
	if !h.selector.IsRegistered(namespace, name) {
		log.Infof("Skipped sts delete event on namespace %s name %s", namespace, name)
		return nil
	}

	defer h.cleanLastSize(namespace, name)
	return h.controller.UpdatePoolSize(ctx, namespace, name, 0)
}

func (h *Handler) hasLastSizeVariation(ss *api.StatefulSet) bool {
	k := ss.Namespace + "/" + ss.Name
	h.mutex.Lock()
	defer func() {
		h.lastSizeIndex[k] = int(*ss.Spec.Replicas)
		h.mutex.Unlock()
	}()

	if _, ok := h.lastSizeIndex[k]; !ok {
		h.lastSizeIndex[k] = 0
	}

	return int(*ss.Spec.Replicas) != h.lastSizeIndex[k]
}

func (h *Handler) cleanLastSize(namespace, name string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	delete(h.lastSizeIndex, namespace+"/"+name)
}
