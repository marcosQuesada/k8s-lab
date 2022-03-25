package crd

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/marcosQuesada/k8s-lab/services/config-reloader-controller/internal/infra/k8s/crd/apis/configmappodrefresher/v1alpha1"
	log "github.com/sirupsen/logrus"
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
	cm := obj.(*v1alpha1.ConfigMapPodRefresher)
	spew.Dump(cm)
	log.Infof("Created ConfigMap %s ", cm.Name)
}

// Updated handles configmap updates event
func (h *Handler) Updated(ctx context.Context, new, old runtime.Object) {
	cm := new.(*v1alpha1.ConfigMapPodRefresher)
	spew.Dump(cm)
	log.Infof("Deleted ConfigMap %s", cm.Name)
}

// Deleted handles configmap deletion event
func (h *Handler) Deleted(ctx context.Context, obj runtime.Object) {
	cm := obj.(*v1alpha1.ConfigMapPodRefresher)
	spew.Dump(cm)
	log.Infof("Deleted ConfigMap %s", cm.Name)
}
