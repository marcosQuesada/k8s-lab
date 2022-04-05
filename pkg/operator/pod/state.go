package pod

import (
	"errors"
	log "github.com/sirupsen/logrus"
	api "k8s.io/api/core/v1"
	"strconv"
	"strings"
)

var ErrBadStatefulSetPodName = errors.New("malformed pod name, expected statefulset pattern")

func IsReadyPod(pod *api.Pod) bool {
	if IsHostNetworked(pod) {
		log.WithField("pod", pod.Name).Debug("Pod is host networked.")
		return false
	} else if !HasIPAddress(pod) {
		log.WithField("pod", pod.Name).Debug("Pod does not have an IP address.")
		return false
	} else if !IsScheduled(pod) {
		log.WithField("pod", pod.Name).Debug("Pod is not scheduled.")
		return false
	} else if !IsRunning(pod) {
		log.WithField("pod", pod.Name).Debug("Pod is not running.")
		return false
	}
	return true
}

func IsScheduled(pod *api.Pod) bool {
	return pod.Spec.NodeName != ""
}

func IsHostNetworked(pod *api.Pod) bool {
	return pod.Spec.HostNetwork
}

func HasIPAddress(pod *api.Pod) bool {
	return pod.Status.PodIP != ""
}

func IsRunning(pod *api.Pod) bool {
	return pod.Status.Phase == "Running"
}

func IsReady(pod *api.Pod) bool {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "True" {
			return true
		}
	}

	return false
}

func HasDeletionTimestamp(pod *api.Pod) bool {
	return pod.ObjectMeta.DeletionTimestamp != nil
}

func IsTerminated(pod *api.Pod) bool {
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Terminated != nil {
			return true
		}
	}
	return false
}

func StatefulSetIndex(pod *api.Pod) (int, error) {
	parts := strings.Split(pod.Name, "-")
	if len(parts) < 2 {
		return 0, ErrBadStatefulSetPodName
	}

	idx := parts[len(parts)-1]
	i, err := strconv.ParseInt(idx, 10, 64)

	return int(i), err
}
