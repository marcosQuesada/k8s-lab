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

// Created handles swarm creation event
func (h *Handler) Created(ctx context.Context, obj runtime.Object) {
	sw := obj.(*v1alpha1.Swarm)

	log.Infof("Created Swarm Namespace %s name %s StatefulSetName %s", sw.Spec.Namespace, sw.Name, sw.Spec.StatefulSetName)
	h.app.Add(sw.Spec.Namespace, sw.Spec.StatefulSetName, sw.Spec.Version, sw.Spec.Workload)
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
	h.app.Update(swn.Spec.Namespace, swn.Spec.StatefulSetName, swn.Spec.Version, swn.Spec.Workload)
}

// Deleted handles statefulset deletion event
func (h *Handler) Deleted(ctx context.Context, obj runtime.Object) {
	sw := obj.(*v1alpha1.Swarm)

	log.Infof("Deleted Swarm Namespace %s name %s watchedLabel %s", sw.Spec.Namespace, sw.Name, sw.Spec.StatefulSetName)
	h.app.Delete(sw.Spec.Namespace, sw.Spec.StatefulSetName) // @TODO: WTF
}

/*** @TODO ONGOING REFACTOR ****/

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
