package statefulset

import (
	"context"
	log "github.com/sirupsen/logrus"
	api "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// PoolController models an ordered set of workers
type PoolController interface {
	Matches(namespace, name string, l map[string]string) bool
	UpdatePool(ctx context.Context, namespace, name string, size int) error
}

// Handler handles statefulset state updates
type Handler struct {
	controller PoolController
}

// NewHandler instantiates statefulset handler
func NewHandler(c PoolController) *Handler {
	return &Handler{
		controller: c,
	}
}

func (h *Handler) Set(ctx context.Context, o runtime.Object) error {
	ss := o.(*api.StatefulSet)
	if !h.controller.Matches(ss.Namespace, ss.Name, ss.Labels) {
		log.Infof("Skiped sts namespace %s name %s labels %v", ss.Namespace, ss.Name, ss.Labels)
		return nil
	}

	log.Infof("Set Statefulset Namespace %s name %s Replicas %d", ss.Namespace, ss.Name, uint64(*ss.Spec.Replicas))

	return h.controller.UpdatePool(ctx, ss.Namespace, ss.Name, int(*ss.Spec.Replicas))
}

func (h *Handler) Remove(ctx context.Context, namespace, name string) error {
	log.Infof("Deleted StatefulSet Namespace %s name %s", namespace, name)

	return h.controller.UpdatePool(ctx, namespace, name, 0)
}
