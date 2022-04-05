package crd

import (
	"context"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/apis/swarm/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"sync"
)

type controller interface {
	Process(ctx context.Context, namespace, name string) error
	Delete(ctx context.Context, namespace, name string) error
}

// Handler handles swarm state updates
type Handler struct {
	controller     controller
	swarmLastIndex map[string]*swarmLastState
	mutex          sync.RWMutex
}

// NewHandler instantiates swarm handler
func NewHandler(c controller) *Handler {
	return &Handler{
		controller:     c,
		swarmLastIndex: map[string]*swarmLastState{},
	}
}

func (h *Handler) Handle(ctx context.Context, o runtime.Object) error {
	sw := o.(*v1alpha1.Swarm)
	log.Infof("Handle Swarm Namespace %s name %s StatefulSet Name %s size %d status %s", sw.Namespace, sw.Name, sw.Spec.StatefulSetName, sw.Spec.Size, sw.Status)

	if !h.hasLastSwarmVariation(sw) {
		return nil
	}

	err := h.controller.Process(ctx, sw.Namespace, sw.Name)
	if err != nil {
		log.Errorf("error processing swarm %s %s ", sw.Namespace, sw.Name)
		return err
	}

	return nil
}

func (h *Handler) HandleDeletion(ctx context.Context, namespace, name string) error {
	log.Infof("HandleDeletion Swarm Namespace %s name %s", namespace, name)

	_ = h.controller.Delete(ctx, namespace, name)
	// @TODO:  check errors out
	h.cleanLastSwarm(namespace, name)
	return nil
}

func (h *Handler) hasLastSwarmVariation(ss *v1alpha1.Swarm) bool {
	k := ss.Namespace + "/" + ss.Name
	h.mutex.Lock()

	defer func() {
		h.swarmLastIndex[k] = &swarmLastState{
			version:         ss.Spec.Version,
			statefulSetName: ss.Spec.StatefulSetName,
			configMapName:   ss.Spec.ConfigMapName,
			workload:        ss.Spec.Workload,
		}
		h.mutex.Unlock()
	}()

	v, ok := h.swarmLastIndex[k]
	if !ok {
		return true
	}

	return v.Equals(ss.Spec.Version, ss.Spec.StatefulSetName, ss.Spec.ConfigMapName, ss.Spec.Workload)
}

func (h *Handler) cleanLastSwarm(namespace, name string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	delete(h.swarmLastIndex, namespace+"/"+name)
}

type swarmLastState struct {
	version         int64
	statefulSetName string
	configMapName   string
	workload        []v1alpha1.Job
}

func (s *swarmLastState) Equals(v int64, stsName, cmName string, wk []v1alpha1.Job) bool {
	if s.version != v {
		return false
	}

	if s.statefulSetName != stsName {
		return false
	}

	if s.configMapName != cmName {
		return false
	}

	if len(s.workload) != len(wk) {
		return false
	}

	jobs := map[v1alpha1.Job]struct{}{}
	for _, job := range s.workload {
		jobs[job] = struct{}{}
	}

	for _, job := range wk {
		if _, ok := jobs[job]; !ok {
			return false
		}
	}

	return true
}
