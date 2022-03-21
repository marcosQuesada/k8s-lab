package operator

import (
	"context"
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"sync"
	"testing"
)

func TestController_RunInitializesAccessingListWatcher(t *testing.T) {
	wi := newFakeWatcher()
	w := &fakeListWatcher{watcher: wi}
	rh := &fakeResourceHandler{}
	eh := &fakeEventHandler{}
	ep := NewEventProcessor(&apiv1.Pod{}, w, eh, rh)
	done := make(chan struct{})
	ep.Run(done)

	if expected, got := 1, w.listed(); expected != got {
		t.Fatalf("handled size does not match, expected %d got %d", expected, got)
	}
	if expected, got := 1, w.watched(); expected != got {
		t.Fatalf("handled size does not match, expected %d got %d", expected, got)
	}
	close(done)
}

var (
	name      = "swarm-worker-0"
	namespace = "swarm"
)

func TestController_HandlePodCreationAndIntrospectIndexer(t *testing.T) {
	wi := newFakeWatcher()
	w := &fakeListWatcher{watcher: wi, namespace: namespace, name: name}
	rh := &fakeResourceHandler{}
	eh := &fakeEventHandler{}
	ep := NewEventProcessor(&apiv1.Pod{}, w, eh, rh)
	done := make(chan struct{})
	defer close(done)
	ep.Run(done)

	pod := getFakePod(namespace, name)
	ev := &event{
		key: fmt.Sprintf("%s/%s", namespace, name),
		obj: &pod,
	}

	if err := ep.Handle(context.Background(), ev); err != nil {
		t.Fatalf("unexpected error handling event, %v", err)
	}

	if expected, got := 1, rh.created; expected != got {
		t.Fatalf("handled size does not match, expected %d got %d", expected, got)
	}

	keys := ep.indexer.ListKeys()
	if expected, got := 1, len(keys); expected != got {
		t.Fatalf("handled size does not match, expected %d got %d", expected, got)
	}

	if expected, got := fmt.Sprintf("%s/%s", namespace, name), keys[0]; expected != got {
		t.Fatalf("keys do not match, expected %s got %s", expected, got)
	}
}

func TestController_HandlePodUpdate(t *testing.T) {
	wi := newFakeWatcher()
	w := &fakeListWatcher{watcher: wi, namespace: namespace, name: name}
	rh := &fakeResourceHandler{}
	eh := &fakeEventHandler{}
	ep := NewEventProcessor(&apiv1.Pod{}, w, eh, rh)
	done := make(chan struct{})
	defer close(done)
	ep.Run(done)

	pod := getFakePod(namespace, name)
	ev := &updateEvent{
		key:    fmt.Sprintf("%s/%s", namespace, name),
		oldObj: &pod,
		newObj: &pod,
	}

	if err := ep.Handle(context.Background(), ev); err != nil {
		t.Fatalf("unexpected error handling event, %v", err)
	}

	if expected, got := 1, rh.updated; expected != got {
		t.Fatalf("handled size does not match, expected %d got %d", expected, got)
	}
}

func TestController_HandlePodDeletion(t *testing.T) {
	wi := newFakeWatcher()
	w := &fakeListWatcher{watcher: wi, namespace: namespace, name: name}
	rh := &fakeResourceHandler{}
	eh := &fakeEventHandler{}
	ep := NewEventProcessor(&apiv1.Pod{}, w, eh, rh)

	pod := getFakePod(namespace, name)
	ev := &updateEvent{
		key:    fmt.Sprintf("%s/%s", namespace, name),
		oldObj: &pod,
		newObj: &pod,
	}

	if err := ep.Handle(context.Background(), ev); err != nil {
		t.Fatalf("unexpected error handling event, %v", err)
	}

	if expected, got := 1, rh.deleted; expected != got {
		t.Fatalf("handled size does not match, expected %d got %d", expected, got)
	}
}

type fakeListWatcher struct {
	name        string
	namespace   string
	watcher     watch.Interface
	sleep       bool
	listCalled  int
	watchCalled int
	mutex       sync.Mutex
}

func (f *fakeListWatcher) List(options metav1.ListOptions) (runtime.Object, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.listCalled++

	pods := &apiv1.PodList{
		Items: []apiv1.Pod{getFakePod(f.namespace, f.name)},
	}
	return pods, nil
}

func (f *fakeListWatcher) Watch(options metav1.ListOptions) (watch.Interface, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.watchCalled++
	return f.watcher, nil
}

func (f *fakeListWatcher) listed() int {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.listCalled
}

func (f *fakeListWatcher) watched() int {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.watchCalled
}

type fakeEventHandler struct {
	totalCreated int
	totalUpdated int
	totalDeleted int
}

func (f *fakeEventHandler) Add(obj interface{}) {
	f.totalCreated++
}

func (f *fakeEventHandler) Update(oldObj, newObj interface{}) {
	f.totalUpdated++
}

func (f *fakeEventHandler) Delete(obj interface{}) {
	f.totalDeleted++
}

type fakeResourceHandler struct {
	created int
	updated int
	deleted int
}

func (f *fakeResourceHandler) Created(ctx context.Context, obj runtime.Object) {
	f.created++
}

func (f *fakeResourceHandler) Updated(ctx context.Context, old, new runtime.Object) {
	f.updated++
}

func (f *fakeResourceHandler) Deleted(ctx context.Context, obj runtime.Object) {
	f.deleted++
}

type fakeWatcher struct {
	events chan watch.Event
}

func newFakeWatcher() *fakeWatcher {
	return &fakeWatcher{
		events: make(chan watch.Event),
	}
}

func (f *fakeWatcher) Stop() {
	close(f.events)
}

func (f *fakeWatcher) ResultChan() <-chan watch.Event {
	return f.events
}

func getFakePod(namespace, name string) apiv1.Pod {
	return apiv1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Name:            "nginx",
					Image:           "nginx",
					ImagePullPolicy: "Always",
				},
			},
			RestartPolicy: apiv1.RestartPolicyNever,
		},
	}
}
