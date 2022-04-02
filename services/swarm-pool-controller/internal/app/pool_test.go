package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/marcosQuesada/k8s-lab/pkg/config"
	"sync"
	"sync/atomic"
	"testing"
)

var version int64 = 1
var namespace = "default"

func TestOnGetAllWorkersReturnsAnOrderedListOfWorkers(t *testing.T) {
	asg := &fakeAssigner{}
	call := &fakeCaller{}
	app := NewWorkerPool(version, asg, call)

	w1Index := 1
	w1Name := "worker-1"
	if added := app.AddWorkerIfNotExists(w1Index, namespace, w1Name); !added {
		t.Fatalf("unable to add worker %s", w1Name)
	}

	w3Index := 3
	w3Name := "worker-3"
	if added := app.AddWorkerIfNotExists(w3Index, namespace, w3Name); !added {
		t.Fatalf("unable to add worker %s", w3Name)
	}

	w2Index := 2
	w2Name := "worker-2"
	if added := app.AddWorkerIfNotExists(w2Index, namespace, w2Name); !added {
		t.Fatalf("unable to add worker %s", w2Name)
	}

	a := app.(*pool)
	wl := a.geAllWorkers()
	if expected, got := 3, len(wl); expected != got {
		t.Fatalf("unexpected worker size, expected %d got %d", expected, got)
	}

	if expected, got := w1Name, wl[0].name; expected != got {
		t.Errorf("unexpected first worker, expected %s got %s", expected, got)
	}

	if expected, got := w3Name, wl[2].name; expected != got {
		t.Errorf("unexpected first worker, expected %s got %s", expected, got)
	}
}

func TestOnUpdateExpectedSizeDetectsVariationAndTriesToAssignKeys(t *testing.T) {
	t.Skip()
	asg := &fakeAssigner{}
	call := &fakeCaller{}
	app := NewWorkerPool(version, asg, call)

	app.UpdateSize(1)
	if !app.AddWorkerIfNotExists(0, namespace, "fakeSlave") {
		t.Fatal("worker addition assertion expected true")
	}

	if atLeast, got := int32(1), atomic.LoadInt32(&call.assigns); got < atLeast {
		t.Errorf("min Workloads calls not match, at least %d got %d", atLeast, got)
	}
}

func TestPool_ItMarksWorkersToNotifyOnScaleUp(t *testing.T) {
	//t.Skip() // @TODO: Update before moving on
	asg := &fakeAssigner{}
	call := &fakeCaller{err: errors.New("fake tiemout")}
	app := NewWorkerPool(version, asg, call)

	app.UpdateSize(1)
	if !app.AddWorkerIfNotExists(0, namespace, "fakeSlave-0") {
		t.Fatal("worker addition assertion expected true")
	}
	app.UpdateSize(2)
	if !app.AddWorkerIfNotExists(1, namespace, "fakeSlave-1") {
		t.Fatal("worker addition assertion expected true")
	}

	if atLeast, got := int32(2), atomic.LoadInt32(&call.assigns); got < atLeast {
		t.Errorf("min Workloads calls not match, at least %d got %d", atLeast, got)
	}
	a := app.(*pool)
	w0, err := a.worker("fakeSlave-0")
	if err != nil {
		t.Fatalf("unable to get worker, error %v", err)
	}
	if expected, got := NeedsRefresh, w0.state; expected != got {
		t.Fatalf("expected state does not match, expected %s got %s", expected, got)
	}

	w1, err := a.worker("fakeSlave-1")
	if err != nil {
		t.Fatalf("unable to get worker, error %v", err)
	}
	if expected, got := NeedsRefresh, w1.state; expected == got {
		t.Fatalf("expected state does not match, expected %s got %s", expected, got)
	}
}

func TestPool_ItMarksWorkersToNotifyOnScaleDown(t *testing.T) {
	t.Skip()
	asg := &fakeAssigner{}
	call := &fakeCaller{err: errors.New("fake tiemout")}
	app := NewWorkerPool(version, asg, call)

	app.UpdateSize(2)
	if !app.AddWorkerIfNotExists(0, namespace, "fakeSlave-0") {
		t.Fatal("worker addition assertion expected true")
	}
	if !app.AddWorkerIfNotExists(1, namespace, "fakeSlave-1") {
		t.Fatal("worker addition assertion expected true")
	}

	app.UpdateSize(1)
	app.RemoveWorkerByName(namespace, "fakeSlave-1")

	if atLeast, got := int32(2), atomic.LoadInt32(&call.assigns); got < atLeast {
		t.Errorf("min Workloads calls not match, at least %d got %d", atLeast, got)
	}

	a := app.(*pool)
	w0, err := a.worker("fakeSlave-0")
	if err != nil {
		t.Fatalf("unable to get worker, error %v", err)
	}
	if expected, got := NeedsRefresh, w0.state; expected != got {
		t.Fatalf("expected state does not match, expected %s got %s", expected, got)
	}

}

type fakeAssigner struct {
	balanceRequests int32
}

func (a *fakeAssigner) BalanceWorkload(totalWorkers int, version int64) error {
	atomic.AddInt32(&a.balanceRequests, 1)
	return nil
}

func (a *fakeAssigner) Workloads() *config.Workloads {
	return &config.Workloads{
		Workloads: map[string]*config.Workload{"fake_0": &config.Workload{}},
		Version:   1,
	}
}

type fakeCaller struct { // @TODO: Change it!
	assigns     int32
	err         error
	assignation *config.Workloads
	mutex       sync.RWMutex
}

func (f *fakeCaller) Assign(ctx context.Context, w *config.Workloads) (err error) {
	atomic.AddInt32(&f.assigns, 1)
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.assignation = w

	return nil
}

func (f *fakeCaller) Assignation(ctx context.Context, w *worker) (*config.Workload, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	v, ok := f.assignation.Workloads[w.name]
	if !ok {
		return nil, fmt.Errorf("index %d not found", w.index)
	}
	return v, nil
}

func (f *fakeCaller) RestartWorker(ctx context.Context, namespace, name string) error {
	return nil
}
