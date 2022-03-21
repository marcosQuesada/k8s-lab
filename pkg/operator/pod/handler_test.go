package pod

import (
	api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestParsePodIndexFromStatefulsetPodName(t *testing.T) {
	podName := "swarm-worker-0"

	p := &api.Pod{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{Name: podName},
		Spec:       api.PodSpec{},
		Status:     api.PodStatus{},
	}

	idx, err := podIndex(p)
	if err != nil {
		t.Fatalf("unable to parse pod index %v", err)
	}

	if expected, got := 0, idx; expected != got {
		t.Errorf("unexpected index, expected %d got %d", expected, got)
	}
}
