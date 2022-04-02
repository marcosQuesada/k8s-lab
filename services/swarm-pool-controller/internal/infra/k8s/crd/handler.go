package crd

import (
	"context"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/apis/swarm/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
)

type App interface {
	Add(namespace, label string, version int64, workloads []v1alpha1.Job)
	Delete(namespace, label string)
}

type AppCtl interface {
	Process(ctx context.Context, namespace, name string) error
}

// Handler handles swarm state updates
type Handler struct {
	app        App
	controller AppCtl
}

// NewHandler instantiates swarm handler
func NewHandler(app App, c AppCtl) *Handler {
	return &Handler{app: app, controller: c}
}

func (h *Handler) Set(ctx context.Context, o runtime.Object) error {
	sw := o.(*v1alpha1.Swarm)

	log.Infof("Set Swarm Namespace %s name %s StatefulSet Name %s size %d status %s", sw.Spec.Namespace, sw.Name, sw.Spec.StatefulSetName, sw.Spec.Size, sw.Status)

	err := h.controller.Process(ctx, sw.Namespace, sw.Name)
	if err != nil {
		log.Errorf("error processing swarm %s %s ", sw.Namespace, sw.Name)
		return err
	}
	// @TODO:
	/**
	- check swarm version against configmap ,It May exist previously!
	- get statefulset size
	- get pods from statefulset [name]
	- assign state [+ configmap dump]
	- schedule restart/rollingUpdate pods
	*/
	return nil
}

func (h *Handler) Remove(ctx context.Context, namespace, name string) error {
	log.Infof("Remove Swarm Namespace %s name %s", namespace, name)

	return nil
}
