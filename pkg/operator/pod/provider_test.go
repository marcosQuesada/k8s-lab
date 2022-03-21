package pod

import (
	"context"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
	"time"
)

func TestNewProvider_ItDeletesPodToRefreshByNewOne(t *testing.T) {
	name := "swarm-worker-0"
	namespace := "swarm"
	clientSet := fake.NewSimpleClientset(&apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			ResourceVersion: "pod.v1",
		},
	})

	gvr, _ := schema.ParseResourceArg("pods.v1.")
	trk := clientSet.Tracker()
	w, err := trk.Watch(*gvr, namespace)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	p := NewProvider(clientSet, namespace)
	if err := p.RefreshPod(context.Background(), name); err != nil {
		t.Fatalf("unexepcted error refreshing pod %s, got %v", name, err)
	}

	var r watch.Event
	select {
	case r = <-w.ResultChan():
	case <-time.NewTimer(time.Second).C:
		t.Fatal("Timeout waiting result")
	}

	if r.Type != watch.Deleted {
		t.Errorf("unexpected event type, got %T", r.Type)
	}

	pod, ok := r.Object.(*apiv1.Pod)
	if !ok {
		t.Fatalf("unexpected object type, got %T", r.Object)
	}
	if expected, got := name, pod.Name; expected != got {
		t.Errorf("Pod names do not match, expected %s got %s", expected, got)
	}
}
