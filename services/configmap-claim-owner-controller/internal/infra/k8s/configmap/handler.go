package configmap

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Handler handles configmap state updates
type Handler struct {
}

// NewHandler instantiates configmap handler
func NewHandler() *Handler {
	return &Handler{}
}

// Created handles configmap creation event
func (h *Handler) Created(ctx context.Context, obj runtime.Object) {
	cm := obj.(*api.ConfigMap)
	spew.Dump(cm)
	log.Debugf("Created ConfigMap %s ", cm.Name)
}

// Updated handles configmap updates event
func (h *Handler) Updated(ctx context.Context, new, old runtime.Object) {
	cm := new.(*api.ConfigMap)
	spew.Dump(cm)
	log.Debugf("Deleted ConfigMap %s", cm.Name)
}

// Deleted handles configmap deletion event
func (h *Handler) Deleted(ctx context.Context, obj runtime.Object) {
	cm := obj.(*api.ConfigMap)
	spew.Dump(cm)
	log.Debugf("Deleted ConfigMap %s", cm.Name)
}
