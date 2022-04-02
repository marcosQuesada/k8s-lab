package operator

import (
	"container/list"
	"fmt"
	ap "github.com/marcosQuesada/k8s-lab/pkg/config"
	"sync"
	"testing"
	"time"
)

func TestResourceEventHandler_AddPodIncludesItemOnQueue(t *testing.T) {
	name := "swarm-worker-0"
	namespace := "swarm"
	q := newFakeQueue()
	reh := NewResourceEventHandler(q)
	p := getFakePod(namespace, name)
	reh.Add(&p)

	if expected, got := 1, q.Len(); expected != got {
		t.Fatalf("unexpected queue size, expected %d got %d", expected, got)
	}
	e, shut := q.Get()
	if shut {
		t.Fatalf("unexpected type, got %T", e)
	}

	key, ok := e.(string)
	if !ok {
		t.Fatalf("unexpected type, got %T", e)
	}
	if expected, got := fmt.Sprintf("%s/%s", namespace, name), key; expected != got {
		t.Fatalf("pod name does not match, expected %s got %s", expected, got)
	}
}

func TestResourceEventHandler_UpdatePodIncludesItemOnQueue(t *testing.T) {
	name := "swarm-worker-0"
	namespace := "swarm"
	q := newFakeQueue()
	reh := NewResourceEventHandler(q)
	p := getFakePod(namespace, name)
	reh.Update(&p, &p)

	if expected, got := 1, q.Len(); expected != got {
		t.Fatalf("unexpected queue size, expected %d got %d", expected, got)
	}
	e, shut := q.Get()
	if shut {
		t.Fatalf("unexpected type, got %T", e)
	}

	key, ok := e.(string)
	if !ok {
		t.Fatalf("unexpected type, got %T", e)
	}
	if expected, got := fmt.Sprintf("%s/%s", namespace, name), key; expected != got {
		t.Fatalf("pod name does not match, expected %s got %s", expected, got)
	}
}

func TestResourceEventHandler_DeletePodIncludesItemOnQueue(t *testing.T) {
	name := "swarm-worker-0"
	namespace := "swarm"
	q := newFakeQueue()
	reh := NewResourceEventHandler(q)
	p := getFakePod(namespace, name)
	reh.Delete(&p)

	if expected, got := 1, q.Len(); expected != got {
		t.Fatalf("unexpected queue size, expected %d got %d", expected, got)
	}
	e, shut := q.Get()
	if shut {
		t.Fatalf("unexpected type, got %T", e)
	}

	key, ok := e.(string)
	if !ok {
		t.Fatalf("unexpected type, got %T", e)
	}
	if expected, got := fmt.Sprintf("%s/%s", namespace, name), key; expected != got {
		t.Fatalf("pod name does not match, expected %s got %s", expected, got)
	}
}

type fakeQueue struct {
	queue          *list.List
	index          map[interface{}]*list.Element
	queuedTimes    map[interface{}]int
	forgottenTimes map[interface{}]int
	done           bool
	cond           *sync.Cond
}

func newFakeQueue() *fakeQueue {
	return &fakeQueue{
		queue:          list.New(),
		index:          map[interface{}]*list.Element{},
		queuedTimes:    map[interface{}]int{},
		forgottenTimes: map[interface{}]int{},
		cond:           sync.NewCond(&sync.Mutex{}),
	}
}

func (f *fakeQueue) Add(item interface{}) {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()
	f.index[item] = f.queue.PushBack(item)
	f.queuedTimes[item]++
}

func (f *fakeQueue) Len() int {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()
	return f.queue.Len()
}

func (f *fakeQueue) Get() (item interface{}, shutdown bool) {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()
	if f.queue.Len() == 0 && !f.ShuttingDown() {
		f.cond.Wait()
	}
	e := f.queue.Front()
	if e == nil {
		return nil, f.done
	}
	return e.Value, f.done
}

func (f *fakeQueue) Done(item interface{}) {
	f.cond.L.Lock()
	defer f.cond.L.Unlock()

	e, ok := f.index[item]
	if !ok || e == nil {
		return
	}

	f.queue.Remove(e)
	delete(f.index, item)
	f.cond.Signal()
}

func (f *fakeQueue) ShutDown() {
	f.done = true
}

func (f *fakeQueue) ShutDownWithDrain() {
}

func (f *fakeQueue) ShuttingDown() bool {
	return f.done
}

func (f *fakeQueue) AddAfter(item interface{}, duration time.Duration) {
	f.Add(item)
}

func (f *fakeQueue) AddRateLimited(item interface{}) {
	f.Add(item)
}

func (f *fakeQueue) Forget(item interface{}) {
	f.Done(item)

	f.forgottenTimes[item]++
}

func (f *fakeQueue) NumRequeues(item interface{}) int {
	v, ok := f.queuedTimes[item]
	if !ok {
		return 0
	}
	return v
}

func (f *fakeQueue) forgotten(item interface{}) int {
	v, ok := f.forgottenTimes[item]
	if !ok {
		return 0
	}
	return v
}

func TestFakeQueueBehaviourDevelopmentTest(t *testing.T) {
	queue := newFakeQueue()
	queue.Add(ap.Job("fooo_0"))
	queue.Add(ap.Job("fooo_1"))
	queue.Add(ap.Job("fooo_2"))
	queue.Add(ap.Job("fooo_3"))
	queue.Add(ap.Job("fooo_4"))

	if expected, got := 5, queue.Len(); expected != got {
		t.Fatalf("unexpected queue size, expected %d got %d", expected, got)
	}

	k0, _ := queue.Get()
	key0, ok := k0.(ap.Job)
	if !ok {
		t.Fatalf("unexpected key type, got %T", k0)
	}

	if expected, got := "fooo_0", string(key0); expected != got {
		t.Fatalf("unexpected collection key, expected %s got %s", expected, got)
	}
	queue.Done(k0)

	k1, _ := queue.Get()
	queue.Done(k1)

	k2, _ := queue.Get()
	queue.Done(k2)

	if expected, got := 2, queue.Len(); expected != got {
		t.Fatalf("unexpected queue size, expected %d got %d", expected, got)
	}

	k3, _ := queue.Get()
	queue.Done(k3)

	k4, _ := queue.Get()
	queue.Done(k4)

	if expected, got := 0, queue.Len(); expected != got {
		t.Fatalf("unexpected queue size, expected %d got %d", expected, got)
	}

	queue.ShutDown()
	_, shut := queue.Get()
	if !shut {
		t.Error("expected queue on shutdown")
	}
}
