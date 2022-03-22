package statefulset

import (
	"context"
	log "github.com/sirupsen/logrus"
	api "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Pool models an ordered set of workers
type Pool interface {
	UpdateExpectedSize(size int)
}

// Handler handles statefulset state updates
type Handler struct {
	state Pool
}

// NewHandler instantiates statefulset handler
func NewHandler(st Pool) *Handler {
	return &Handler{
		state: st,
	}
}

// Created handles statefulset creation event
func (h *Handler) Created(ctx context.Context, obj runtime.Object) {
	ss := obj.(*api.StatefulSet)
	h.state.UpdateExpectedSize(int(*ss.Spec.Replicas))

	log.Debugf("Created StatefulSet %s replicas %d", ss.Name, uint64(*ss.Spec.Replicas))
}

// Updated handles statefulset updates event
func (h *Handler) Updated(ctx context.Context, new, old runtime.Object) {
	ss := new.(*api.StatefulSet)
	h.state.UpdateExpectedSize(int(*ss.Spec.Replicas))
}

// Deleted handles statefulset deletion event
func (h *Handler) Deleted(ctx context.Context, obj runtime.Object) {
	ss := obj.(*api.StatefulSet)
	h.state.UpdateExpectedSize(0)

	log.Debugf("Deleted StatefulSet %s", ss.Name)
}
