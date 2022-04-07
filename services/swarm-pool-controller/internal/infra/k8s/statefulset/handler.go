package statefulset

import (
	"context"
	log "github.com/sirupsen/logrus"
	api "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// PoolController models an ordered set of workers
type PoolController interface {
	UpdatePoolSize(ctx context.Context, namespace, name string, size int) error
}

// Handler handles statefulSet state updates
type Handler struct {
	controller PoolController
	selector   SelectorStore
}

// NewHandler instantiates statefulset handler
func NewHandler(c PoolController, s SelectorStore) *Handler {
	return &Handler{
		controller: c,
		selector:   s,
	}
}

func (h *Handler) Create(ctx context.Context, o runtime.Object) error {
	ss := o.(*api.StatefulSet)
	if !h.selector.Matches(ss.Namespace, ss.Name, ss.Labels) {
		log.Infof("Skipped sts event namespace %s name %s labels %v", ss.Namespace, ss.Name, ss.Labels)
		return nil
	}

	log.Infof("Create Statefulset Namespace %s name %s Replicas %d", ss.Namespace, ss.Name, uint64(*ss.Spec.Replicas))

	return h.controller.UpdatePoolSize(ctx, ss.Namespace, ss.Name, int(*ss.Spec.Replicas))
}

func (h *Handler) Update(ctx context.Context, o, n runtime.Object) error {
	oss := o.(*api.StatefulSet)
	nss := n.(*api.StatefulSet)
	if !h.selector.Matches(oss.Namespace, oss.Name, oss.Labels) {
		log.Infof("Skipped sts event namespace %s name %s labels %v", oss.Namespace, oss.Name, oss.Labels)
		return nil
	}

	if oss.Spec.Replicas == nss.Spec.Replicas {
		return nil
	}

	log.Infof("Create Statefulset Namespace %s name %s Replicas %d", nss.Namespace, nss.Name, uint64(*nss.Spec.Replicas))

	return h.controller.UpdatePoolSize(ctx, nss.Namespace, nss.Name, int(*nss.Spec.Replicas))
}

func (h *Handler) Delete(ctx context.Context, o runtime.Object) error {
	sts := o.(*api.StatefulSet)
	log.Infof("Deleted StatefulSet Namespace %s name %s", sts.Namespace, sts.Name)
	if !h.selector.IsRegistered(sts.Namespace, sts.Name) {
		log.Infof("Skipped sts delete event on namespace %s name %s", sts.Namespace, sts.Name)
		return nil
	}

	return h.controller.UpdatePoolSize(ctx, sts.Namespace, sts.Name, 0)
}
