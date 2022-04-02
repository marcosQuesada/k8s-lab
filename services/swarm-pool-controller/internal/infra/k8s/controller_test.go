package k8s

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/marcosQuesada/k8s-lab/pkg/operator"
	"github.com/marcosQuesada/k8s-lab/pkg/operator/controller"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/apis/swarm/v1alpha1"
	fake "github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/generated/clientset/versioned/fake"
	swarmInformer "github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/generated/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"log"
	"sync/atomic"
	"testing"
	"time"
)

func TestFooController_Run(t *testing.T) {
	clientset := operator.BuildExternalClient()
	swarmClientset := BuildSwarmExternalClient()
	done := make(chan struct{})

	i := informers.NewSharedInformerFactory(clientset, 0)
	di := i.Apps().V1().Deployments()
	go di.Informer().Run(done)

	swSif := swarmInformer.NewSharedInformerFactory(swarmClientset, 0)
	go swSif.K8slab().V1alpha1().Swarms().Informer().Run(done)

	ctl := NewController(clientset, swarmClientset, di, swSif.K8slab().V1alpha1().Swarms())

	go ctl.Run(1, done)

	time.Sleep(time.Second * 100)
}

func TestFooController_PodStore(t *testing.T) {
	clientset := operator.BuildExternalClient()

	done := make(chan struct{})
	defer close(done)
	i := informers.NewSharedInformerFactory(clientset, 0)

	si := i.Apps().V1().StatefulSets()
	go si.Informer().Run(done)

	pi := i.Core().V1().Pods()
	go pi.Informer().Run(done)

	if !cache.WaitForNamedCacheSync("statefulset", done, pi.Informer().HasSynced, si.Informer().HasSynced) {
		t.Fatal("unable to sync")
	}

	sts, err := si.Lister().StatefulSets("swarm").Get("swarm-worker")
	if err != nil {
		t.Fatalf("unable to sync, error %v", err)
	}

	selector, _ := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	pods, err := pi.Lister().Pods(sts.Namespace).List(selector)
	if err != nil {
		t.Fatalf("unable to sync, error %v", err)
	}
	spew.Dump(pods, err)

	l, err := labels.NewRequirement("app", selection.In, []string{"swarm-worker"})
	if err != nil {
		log.Fatal("Labels requirement must validate successfully")
	}
	sl := labels.NewSelector()
	sl.Add(*l)

	ps, err := pi.Lister().Pods(sts.Namespace).List(selector)
	if err != nil {
		t.Fatalf("unable to sync, error %v", err)
	}
	spew.Dump(ps)

	if !sl.Matches(labels.Set(ps[0].Labels)) {
		t.Fatal("expected equality")
	}

	if !sl.Matches(labels.Set(sts.Labels)) {
		t.Fatal("expected equality")
	}

	stsl, err := si.Lister().StatefulSets("swarm").List(sl)
	if err != nil {
		t.Fatalf("unable to sync, error %v", err)
	}
	spew.Dump(stsl)
}

func TestControllerImplementationWorkflow(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var namespace = "swarm"
	var name = "foo-bar-worker"
	sw := getFakeSwarm(namespace, name)
	eh := &FakeHandler{}
	cl := fake.NewSimpleClientset(sw)
	swSif := swarmInformer.NewSharedInformerFactory(cl, 0)
	inf := swSif.K8slab().V1alpha1().Swarms().Informer()

	ctl := controller.New(eh, inf, "Swarm")
	go ctl.Run(ctx)

	if err := inf.GetIndexer().Add(sw); err != nil {
		t.Fatalf("unable to add entry to indexer, error %v", err)
	}

	time.Sleep(time.Second)

	if expected, got := 1, eh.created(); expected < int(got) {
		t.Errorf("calls do not match, expected %d got %d", expected, got)
	}
	if expected, got := 0, eh.deleted(); expected != int(got) {
		t.Errorf("calls do not match, expected %d got %d", expected, got)
	}
}

type FakeHandler struct {
	totalCreated int32
	totalDeleted int32
}

func (f *FakeHandler) Set(ctx context.Context, o runtime.Object) error {
	atomic.AddInt32(&f.totalCreated, 1)
	return nil
}

func (f *FakeHandler) Remove(ctx context.Context, namespace, name string) error {
	atomic.AddInt32(&f.totalDeleted, 1)
	return nil
}
func (f *FakeHandler) created() int32 {
	return atomic.LoadInt32(&f.totalCreated)
}

func (f *FakeHandler) deleted() int32 {
	return atomic.LoadInt32(&f.totalDeleted)
}
func getFakeSwarm(namespace, name string) *v1alpha1.Swarm {
	return &v1alpha1.Swarm{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Swarm",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.SwarmSpec{
			Version: 1,
			Size:    2,
			Members: []v1alpha1.Worker{{
				Name: "foobar_0",
				Jobs: []v1alpha1.Job{
					"stream:xxctv12:updated",
					"stream:xxctv13:updated",
					"stream:xxctv14:updated",
					"stream:yxctv1:updated",
					"stream:yxctv2:updated",
					"stream:yxctv3:updated",
					"stream:xabcn0:updated",
					"stream:xacb01:updated",
					"stream:xacb02:updated",
					"stream:xacb03:updated",
					"stream:xacb04:updated",
					"stream:sportnews0:updated",
					"stream:cars:new",
				},
				State:     v1alpha1.Status{Phase: "FAKED"},
				CreatedAt: time.Now().Unix(),
			},
			},
		},
	}
}
