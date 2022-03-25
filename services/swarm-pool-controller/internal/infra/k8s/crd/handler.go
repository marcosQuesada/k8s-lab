package crd

import (
	"context"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/apis/swarm/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
	"unicode"
)

type App interface {
	Add(namespace, label string, version int64, workloads []v1alpha1.Job)
	Update(namespace, label string, version int64, workloads []v1alpha1.Job)
	Delete(namespace, label string)
}

// Handler handles swarm state updates
type Handler struct {
	app App
}

// NewHandler instantiates swarm handler
func NewHandler(app App) *Handler {
	return &Handler{app: app}
}

// Created handles swarm creation event
func (h *Handler) Created(ctx context.Context, obj runtime.Object) {
	sw := obj.(*v1alpha1.Swarm)

	log.Infof("Created Swarm Namespace %s name %s watchedLabel %s", sw.Spec.Namespace, sw.Name, sw.Spec.WatchedLabel)
	h.app.Add(sw.Spec.Namespace, sw.Spec.WatchedLabel, sw.Spec.Version, sw.Spec.Workload)
}

// Updated handles swarm updates event
func (h *Handler) Updated(ctx context.Context, new, old runtime.Object) {
	swn := new.(*v1alpha1.Swarm)
	swo := old.(*v1alpha1.Swarm)

	diff := cmp.Diff(swo, swn)
	cleanDiff := strings.TrimFunc(diff, func(r rune) bool {
		return !unicode.IsGraphic(r)
	})
	fmt.Println("UPDATE Swarm diff: ", cleanDiff)
	h.app.Update(swn.Spec.Namespace, swn.Spec.WatchedLabel, swn.Spec.Version, swn.Spec.Workload)
}

// Deleted handles statefulset deletion event
func (h *Handler) Deleted(ctx context.Context, obj runtime.Object) {
	sw := obj.(*v1alpha1.Swarm)

	log.Infof("Deleted Swarm Namespace %s name %s watchedLabel %s", sw.Spec.Namespace, sw.Name, sw.Spec.WatchedLabel)
	h.app.Delete(sw.Spec.Namespace, sw.Spec.WatchedLabel)
}
