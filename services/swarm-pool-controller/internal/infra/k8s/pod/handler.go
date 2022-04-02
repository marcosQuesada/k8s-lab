package pod

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"strconv"
	"strings"
)

var errBadStatefulSetPodName = errors.New("malformed pod name, expected statefulset pattern")

// Pool models an ordered set of workers
type Pool interface {
	AddWorkerIfNotExists(idx int, namespace, name string) bool
	RemoveWorkerByName(name string)
}

// Handler process Pod state variations
type Handler struct {
	state Pool
}

// NewHandler instantiates pod handler
func NewHandler(st Pool) *Handler {
	return &Handler{
		state: st,
	}
}

// Set on pod creation handler
func (h *Handler) Set(ctx context.Context, obj runtime.Object) {
	pod := obj.(*api.Pod)

	if !isReadyPod(pod) {
		return
	}

	// Quick deletion detection
	if hasDeletionTimestamp(pod) || isTerminated(pod) {
		h.Remove(ctx, pod.Namespace, pod.Name)
		return
	}
	idx, err := podIndex(pod)
	if err != nil {
		log.Errorf("unable to get pod index %v", err)
		return
	}

	if !h.state.AddWorkerIfNotExists(idx, pod.Namespace, pod.Name) {
		return
	}

	log.Debugf("Created Pod %s IP %s", pod.Name, pod.Status.PodIP)
}

// Remove on pod deleted handler
func (h *Handler) Remove(_ context.Context, namespace, name string) {
	log.Debugf("Deleted POD %s/%s", namespace, name)

	h.state.RemoveWorkerByName(name)
}

func isReadyPod(pod *api.Pod) bool {
	if isHostNetworked(pod) {
		log.WithField("pod", pod.Name).Debug("Pod is host networked.")
		return false
	} else if !hasIPAddress(pod) {
		log.WithField("pod", pod.Name).Debug("Pod does not have an IP address.")
		return false
	} else if !isScheduled(pod) {
		log.WithField("pod", pod.Name).Debug("Pod is not scheduled.")
		return false
	} else if !isRunning(pod) {
		log.WithField("pod", pod.Name).Debug("Pod is not running.")
		return false
	}
	return true
}

func isScheduled(pod *api.Pod) bool {
	return pod.Spec.NodeName != ""
}

func isHostNetworked(pod *api.Pod) bool {
	return pod.Spec.HostNetwork
}

func hasIPAddress(pod *api.Pod) bool {
	return pod.Status.PodIP != ""
}

func isRunning(pod *api.Pod) bool {
	return pod.Status.Phase == "Running"
}

func isReady(pod *api.Pod) bool {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "True" {
			return true
		}
	}

	return false
}

func hasDeletionTimestamp(pod *api.Pod) bool {
	return pod.ObjectMeta.DeletionTimestamp != nil
}

func isTerminated(pod *api.Pod) bool {
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Terminated != nil {
			return true
		}
	}
	return false
}

func podIndex(pod *api.Pod) (int, error) {
	parts := strings.Split(pod.Name, "-")
	if len(parts) < 2 {
		return 0, errBadStatefulSetPodName
	}

	idx := parts[len(parts)-1]
	i, err := strconv.ParseInt(idx, 10, 64)

	return int(i), err
}
