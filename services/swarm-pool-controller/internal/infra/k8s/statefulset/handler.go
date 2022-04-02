package statefulset

import (
	"context"
	log "github.com/sirupsen/logrus"
	api "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Pool models an ordered set of workers
type Pool interface {
	UpdateSize(size int)
	Size() int
}

type AppCtl interface {
	Matches(namespace, name string, l map[string]string) bool
	UpdatePool(ctx context.Context, namespace, name string) error
}

// Handler handles statefulset state updates
type Handler struct {
	state      Pool
	controller AppCtl
}

// NewHandler instantiates statefulset handler
func NewHandler(st Pool, c AppCtl) *Handler {
	return &Handler{
		state:      st,
		controller: c,
	}
}

// Created handles statefulset creation event
func (h *Handler) Created(ctx context.Context, obj runtime.Object) {
	ss := obj.(*api.StatefulSet)
	h.state.UpdateSize(int(*ss.Spec.Replicas))

	log.Debugf("Created StatefulSet %s replicas %d", ss.Name, uint64(*ss.Spec.Replicas))
}

// Updated handles statefulset updates event
func (h *Handler) Updated(ctx context.Context, new, old runtime.Object) {
	ss := new.(*api.StatefulSet)
	h.state.UpdateSize(int(*ss.Spec.Replicas))
}

// Deleted handles statefulset deletion event
func (h *Handler) Deleted(ctx context.Context, obj runtime.Object) {
	ss := obj.(*api.StatefulSet)
	h.state.UpdateSize(0)

	log.Debugf("Deleted StatefulSet %s", ss.Name)
}

func (h *Handler) Set(ctx context.Context, o runtime.Object) error {
	ss := o.(*api.StatefulSet)
	if !h.controller.Matches(ss.Namespace, ss.Name, ss.Labels) {
		log.Infof("Skiped sts namespace %s name %s labels %v", ss.Namespace, ss.Name, ss.Labels)
		return nil
	}

	if int(*ss.Spec.Replicas) == h.state.Size() {
		return nil
	}
	h.state.UpdateSize(int(*ss.Spec.Replicas))

	log.Infof("Set Statefulset Namespace %s name %s Replicas %d", ss.Namespace, ss.Name, uint64(*ss.Spec.Replicas))

	return h.controller.UpdatePool(ctx, ss.Namespace, ss.Name)
}

func (h *Handler) Remove(ctx context.Context, namespace, name string) error {
	log.Infof("Deleted StatefulSet Namespace %s name %s", namespace, name)

	return h.controller.UpdatePool(ctx, namespace, name)
}
