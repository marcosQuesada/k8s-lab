package operator

import (
	"context"
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"sync/atomic"
	"testing"
	"time"
)

func TestController_ItGetsCreatedOnListeningPodsWithPodAddition(t *testing.T) {
	namespace := "default"
	name := "foo"
	eh := &fakeHandler{}
	p := getFakePod(namespace, name)
	cl := fake.NewSimpleClientset(p)
	i := informers.NewSharedInformerFactory(cl, 0)
	pi := i.Core().V1().Pods()
	ctl := New(eh, pi.Informer(), "Pod")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go ctl.Run(ctx)

	if err := pi.Informer().GetIndexer().Add(p); err != nil {
		t.Fatalf("unable to add entry to indexer,error %v", err)
	}

	keys := pi.Informer().GetIndexer().ListKeys()
	if expected, got := 1, len(keys); expected != got {
		t.Fatalf("handled size does not match, expected %d got %d", expected, got)
	}

	if expected, got := fmt.Sprintf("%s/%s", namespace, name), keys[0]; expected != got {
		t.Fatalf("keys do not match, expected %s got %s", expected, got)
	}

	// informer runner needs time @TODO: think on a real synced solution
	time.Sleep(time.Second)

	if expected, got := 1, eh.created(); expected != int(got) {
		t.Errorf("calls do not match, expected %d got %d", expected, got)
	}
	if expected, got := 0, eh.deleted(); expected != int(got) {
		t.Errorf("calls do not match, expected %d got %d", expected, got)
	}

	gvr := schema.GroupVersionResource{Resource: "pods"}
	gvk := schema.GroupVersionKind{Group: "apps", Version: "v1"}
	actions := []core.Action{
		core.NewListAction(gvr, gvk, namespace, metav1.ListOptions{}),
		core.NewWatchAction(gvr, namespace, metav1.ListOptions{}),
	}

	clActions := cl.Actions()
	if expected, got := 2, len(clActions); expected != got {
		t.Fatalf("unexpected total actions executed, expected %d got %d", expected, got)
	}

	for i, action := range clActions {
		if len(actions) < i+1 {
			t.Errorf("%d unexpected actions: %+v", len(actions)-len(clActions), actions[i:])
			break
		}

		expectedAction := actions[i]
		if !(expectedAction.Matches(action.GetVerb(), action.GetResource().Resource) && action.GetSubresource() == expectedAction.GetSubresource()) {
			t.Errorf("Expected %#v got %#v", expectedAction, action)
			continue
		}
	}

	if len(actions) > len(clActions) {
		t.Errorf("%d additional expected actions:%+v", len(actions)-len(clActions), actions[len(clActions):])
	}
}

func TestController_ItGetsDeletedOnListeningPodsWithPodAdditionWithoutBeingPreloadedInTheIndexer(t *testing.T) {
	namespace := "default"
	name := "foo"
	eh := &fakeHandler{}
	cl := fake.NewSimpleClientset()
	i := informers.NewSharedInformerFactory(cl, 0)
	pi := i.Core().V1().Pods()
	ctl := New(eh, pi.Informer(), "Pod")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go ctl.Run(ctx)

	p := getFakePod(namespace, name)
	if err := pi.Informer().GetIndexer().Add(p); err != nil {
		t.Fatalf("unable to add entry to indexer,error %v", err)
	}

	keys := pi.Informer().GetIndexer().ListKeys()
	if expected, got := 1, len(keys); expected != got {
		t.Fatalf("handled size does not match, expected %d got %d", expected, got)
	}

	// informer runner needs time @TODO: think on a real synced solution
	time.Sleep(time.Second)

	if expected, got := 0, eh.created(); expected != int(got) {
		t.Errorf("calls do not match, expected %d got %d", expected, got)
	}
	if expected, got := 1, eh.deleted(); expected != int(got) {
		t.Errorf("calls do not match, expected %d got %d", expected, got)
	}
}

type fakeHandler struct {
	totalCreated int32
	totalDeleted int32
}

func (f *fakeHandler) Handle(ctx context.Context, o runtime.Object) error {
	atomic.AddInt32(&f.totalCreated, 1)
	return nil
}

func (f *fakeHandler) HandleDeletion(ctx context.Context, namespace, name string) error {
	atomic.AddInt32(&f.totalDeleted, 1)
	return nil
}
func (f *fakeHandler) created() int32 {
	return atomic.LoadInt32(&f.totalCreated)
}

func (f *fakeHandler) deleted() int32 {
	return atomic.LoadInt32(&f.totalDeleted)
}

func getFakePod(namespace, name string) *apiv1.Pod {
	return &apiv1.Pod{
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
