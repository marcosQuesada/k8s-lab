package operator

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func TestResourceEventHandler_AddPodIncludesItemOnQueue(t *testing.T) {
	name := "swarm-worker-0"
	namespace := "swarm"
	s := &fakeSelector{}
	q := newFakeQueue()
	reh := NewResourceEventHandler(s, q)
	p := getFakePod(namespace, name)
	reh.Add(&p)

	if expected, got := 1, q.Len(); expected != got {
		t.Fatalf("unexpected queue size, expected %d got %d", expected, got)
	}
	e, _ := q.Get()

	ev, ok := e.(*event)
	if !ok {
		t.Fatalf("unexpected type, got %T", e)
	}

	pod, ok := ev.obj.(*apiv1.Pod)
	if !ok {
		t.Fatalf("unexpected type, got %T", ev.obj)
	}
	if expected, got := name, pod.Name; expected != got {
		t.Fatalf("pod name does not match, expected %s got %s", expected, got)
	}
}

func TestResourceEventHandler_UpdatePodIncludesItemOnQueue(t *testing.T) {
	name := "swarm-worker-0"
	namespace := "swarm"
	s := &fakeSelector{}
	q := newFakeQueue()
	reh := NewResourceEventHandler(s, q)
	p := getFakePod(namespace, name)
	reh.Update(&p, &p)

	if expected, got := 1, q.Len(); expected != got {
		t.Fatalf("unexpected queue size, expected %d got %d", expected, got)
	}
	e, _ := q.Get()

	ev, ok := e.(*updateEvent)
	if !ok {
		t.Fatalf("unexpected type, got %T", e)
	}

	pod, ok := ev.newObj.(*apiv1.Pod)
	if !ok {
		t.Fatalf("unexpected type, got %T", ev.newObj)
	}
	if expected, got := name, pod.Name; expected != got {
		t.Fatalf("pod name does not match, expected %s got %s", expected, got)
	}
}

func TestResourceEventHandler_DeletePodIncludesItemOnQueue(t *testing.T) {
	name := "swarm-worker-0"
	namespace := "swarm"
	s := &fakeSelector{}
	q := newFakeQueue()
	reh := NewResourceEventHandler(s, q)
	p := getFakePod(namespace, name)
	reh.Delete(&p)

	if expected, got := 1, q.Len(); expected != got {
		t.Fatalf("unexpected queue size, expected %d got %d", expected, got)
	}
	e, _ := q.Get()

	ev, ok := e.(*event)
	if !ok {
		t.Fatalf("unexpected type, got %T", e)
	}

	pod, ok := ev.obj.(*apiv1.Pod)
	if !ok {
		t.Fatalf("unexpected type, got %T", ev.obj)
	}
	if expected, got := name, pod.Name; expected != got {
		t.Fatalf("pod name does not match, expected %s got %s", expected, got)
	}
}

type fakeSelector struct{}

func (f fakeSelector) Validate(object runtime.Object) error {
	return nil
}
