package operator

import (
	"container/list"
	"context"
	"errors"
	ap "github.com/marcosQuesada/k8s-lab/pkg/config"
	apiv1 "k8s.io/api/core/v1"
	"sync"
	"testing"
	"time"
)

func TestController_ManualRunProcessIngestedEventWithSuccess(t *testing.T) {
	feh := &fakeEventProcessor{}
	q := newFakeQueue()
	c := &consumer{
		processor: feh,
		queue:     q,
	}

	ev := &event{
		key: "foo_0",
		obj: &apiv1.Pod{},
	}
	q.Add(ev)

	c.processNextItem()

	sn := feh.snapshot()
	if expected, got := 1, len(sn); expected != got {
		t.Fatalf("handled size does not match, expected %d got %d", expected, got)
	}
	if expected, got := 1, q.forgotten(ev); expected != got {
		t.Fatalf("handled size does not match, expected %d got %d", expected, got)
	}
}

func TestController_ManualRunProcessIngestedEventWithErrorRetriesAgain(t *testing.T) {
	feh := &fakeEventProcessor{err: errors.New("foo error")}
	q := newFakeQueue()
	c := &consumer{
		processor: feh,
		queue:     q,
	}

	key := "foo_0"
	q.Add(&event{
		key: key,
		obj: &apiv1.Pod{},
	})

	c.processNextItem()
	c.processNextItem()
	c.processNextItem()

	sn := feh.snapshot()
	if expected, got := 3, len(sn); expected != got {
		t.Fatalf("handled size does not match, expected %d got %d", expected, got)
	}
}
func TestController_ManualRunProcessIngestedEventWithErrorRetriesAgainUntilMaxRetriesAndDiscardPacket(t *testing.T) {
	feh := &fakeEventProcessor{err: errors.New("foo error")}
	q := newFakeQueue()
	c := &consumer{
		processor: feh,
		queue:     q,
	}

	key := "foo_0"
	ev := &event{
		key: key,
		obj: &apiv1.Pod{},
	}
	q.Add(ev)

	c.processNextItem()
	c.processNextItem()
	c.processNextItem()
	c.processNextItem()
	c.processNextItem()

	sn := feh.snapshot()
	if expected, got := 5, len(sn); expected != got {
		t.Fatalf("handled size does not match, expected %d got %d", expected, got)
	}

	if expected, got := 5, q.NumRequeues(ev); expected != got {
		t.Fatalf("handled size does not match, expected %d got %d", expected, got)
	}
	if expected, got := 1, q.forgotten(ev); expected != got {
		t.Fatalf("handled size does not match, expected %d got %d", expected, got)
	}
}

func TestController_RunProcessIngestedEventFromConciliationLoop(t *testing.T) {
	feh := &fakeEventProcessor{wg: &sync.WaitGroup{}}
	q := newFakeQueue()
	c := &consumer{
		processor:             feh,
		queue:                 q,
		conciliationFrequency: time.Millisecond * 50,
	}

	key := "foo_0"
	q.Add(&event{
		key: key,
		obj: &apiv1.Pod{},
	})
	feh.wg.Add(1)
	done := make(chan struct{})
	go c.Run(done)

	if err := waitUntilTimeout(feh.wg, time.Millisecond*100); err != nil {
		t.Fatal("timeout waiting to handle event")
	}

	sn := feh.snapshot()
	if expected, got := 1, len(sn); expected != got {
		t.Fatalf("handled size does not match, expected %d got %d", expected, got)
	}
}

func waitUntilTimeout(wg *sync.WaitGroup, timeout time.Duration) error {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return nil
	case <-time.After(timeout):
		return errors.New("timeout")
	}
}

type fakeEventProcessor struct {
	mutex   sync.RWMutex
	handled []Event
	err     error
	wg      *sync.WaitGroup
}

func (f *fakeEventProcessor) Run(stopCh chan struct{}) {}

func (f *fakeEventProcessor) Handle(ctx context.Context, ev Event) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.handled = append(f.handled, ev)

	if f.wg != nil {
		f.wg.Done()
	}
	return f.err
}

func (f *fakeEventProcessor) snapshot() []Event {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return f.handled
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
