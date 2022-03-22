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

// Handler handles swarm state updates
type Handler struct {
}

// NewHandler instantiates swarm handler
func NewHandler() *Handler {
	return &Handler{}
}

// Created handles swarm creation event
func (h *Handler) Created(ctx context.Context, obj runtime.Object) {
	sw := obj.(*v1alpha1.Swarm)

	log.Infof("Created Swarm %v", sw)
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
}

// Deleted handles statefulset deletion event
func (h *Handler) Deleted(ctx context.Context, obj runtime.Object) {
	sw := obj.(*v1alpha1.Swarm)

	log.Infof("Deleted Swarm %v", sw)
}
