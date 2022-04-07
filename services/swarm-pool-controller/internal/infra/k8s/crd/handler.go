package crd

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/apis/swarm/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"sync"
)

type controller interface {
	Create(ctx context.Context, namespace, name string) error
	Update(ctx context.Context, namespace, name string) error
	Delete(ctx context.Context, namespace, name string) error
}

// Handler handles swarm state updates
type Handler struct {
	controller controller
	mutex      sync.RWMutex
}

// NewHandler instantiates swarm handler
func NewHandler(c controller) *Handler {
	return &Handler{
		controller: c,
	}
}

func (h *Handler) Create(ctx context.Context, o runtime.Object) error {
	sw := o.(*v1alpha1.Swarm)
	log.Infof("Create Swarm Namespace %s name %s StatefulSet Name %s size %d status %s", sw.Namespace, sw.Name, sw.Spec.StatefulSetName, sw.Spec.Size, sw.Status)

	if err := h.controller.Create(ctx, sw.Namespace, sw.Name); err != nil {
		return fmt.Errorf("unable to process swarm %s %s error %v", sw.Namespace, sw.Name, err)
	}

	return nil
}

func (h *Handler) Update(ctx context.Context, o, n runtime.Object) error {
	osw := o.(*v1alpha1.Swarm)
	nsw := n.(*v1alpha1.Swarm)
	log.Infof("Create Swarm Namespace %s name %s StatefulSet Name %s size %d status %s", osw.Namespace, osw.Name, osw.Spec.StatefulSetName, osw.Spec.Size, osw.Status)

	if Equals(osw, nsw) {
		return nil
	}

	if err := h.controller.Update(ctx, nsw.Namespace, nsw.Name); err != nil {
		return fmt.Errorf("unable to update swarm %s %s error %v", nsw.Namespace, nsw.Name, err)
	}

	return nil
}

func (h *Handler) Delete(ctx context.Context, o runtime.Object) error {
	sw := o.(*v1alpha1.Swarm)
	log.Infof("Delete Swarm Namespace %s name %s", sw.Namespace, sw.Name)

	if err := h.controller.Delete(ctx, sw.Namespace, sw.Name); err != nil {
		return fmt.Errorf("unable to delete swarm %s %s error %v", sw.Namespace, sw.Name, err)
	}

	return nil
}

func Equals(o, n *v1alpha1.Swarm) bool {
	if o.Spec.StatefulSetName != n.Spec.StatefulSetName {
		return false
	}

	if o.Spec.ConfigMapName != n.Spec.ConfigMapName {
		return false
	}

	if len(o.Spec.Workload) != len(n.Spec.Workload) {
		return false
	}

	jobs := map[v1alpha1.Job]struct{}{}
	for _, job := range o.Spec.Workload {
		jobs[job] = struct{}{}
	}

	for _, job := range n.Spec.Workload {
		if _, ok := jobs[job]; !ok {
			return false
		}
	}

	return true
}
