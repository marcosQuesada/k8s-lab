package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/marcosQuesada/k8s-lab/pkg/config"
	"net"
	"sync"
	"sync/atomic"
	"testing"
)

// @TODO: Increase testing coverage, include conciliation loop and state assertions

func TestOnGetAllWorkersReturnsAnOrderedListOfWorkers(t *testing.T) {
	asg := &fakeAssigner{}
	call := &fakeCaller{}
	app := NewWorkerPool(asg, call)
	// avoid any noise from conciliation loop
	app.Terminate()

	w1Index := 1
	w1IP := net.ParseIP("8.8.8.1")
	w1Name := "worker-1"
	if added := app.AddWorkerIfNotExists(w1Index, w1Name, w1IP); !added {
		t.Fatalf("unable to add worker %s error %s", w1Name, w1IP.String())
	}

	w3Index := 3
	w3IP := net.ParseIP("8.8.8.3")
	w3Name := "worker-3"
	if added := app.AddWorkerIfNotExists(w3Index, w3Name, w3IP); !added {
		t.Fatalf("unable to add worker %s error %s", w3Name, w3IP.String())
	}

	w2Index := 2
	w2IP := net.ParseIP("8.8.8.2")
	w2Name := "worker-2"
	if added := app.AddWorkerIfNotExists(w2Index, w2Name, w2IP); !added {
		t.Fatalf("unable to add worker %s error %s", w2Name, w2IP.String())
	}

	wl := app.geAllWorkers()
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
	asg := &fakeAssigner{}
	call := &fakeCaller{}
	app := NewWorkerPool(asg, call)
	defer app.Terminate()

	app.UpdateSize(1)
	if !app.AddWorkerIfNotExists(0, "fakeSlave", net.ParseIP("127.0.0.1")) {
		t.Fatal("worker addition assertion expected true")
	}

	if atLeast, got := int32(1), atomic.LoadInt32(&call.assigns); got < atLeast {
		t.Errorf("min Workloads calls not match, at least %d got %d", atLeast, got)
	}
}

func TestPool_ItMarksWorkersToNotifyOnScaleUp(t *testing.T) {
	asg := &fakeAssigner{}
	call := &fakeCaller{err: errors.New("fake tiemout")}
	app := NewWorkerPool(asg, call)
	defer app.Terminate()

	app.UpdateSize(1)
	if !app.AddWorkerIfNotExists(0, "fakeSlave-0", net.ParseIP("127.0.0.1")) {
		t.Fatal("worker addition assertion expected true")
	}
	app.UpdateSize(2)
	if !app.AddWorkerIfNotExists(1, "fakeSlave-1", net.ParseIP("127.0.0.2")) {
		t.Fatal("worker addition assertion expected true")
	}

	if atLeast, got := int32(2), atomic.LoadInt32(&call.assigns); got < atLeast {
		t.Errorf("min Workloads calls not match, at least %d got %d", atLeast, got)
	}
	w0, err := app.worker("fakeSlave-0")
	if err != nil {
		t.Fatalf("unable to get worker, error %v", err)
	}
	if expected, got := NeedsRefresh, w0.state; expected != got {
		t.Fatalf("expected state does not match, expected %s got %s", expected, got)
	}

	w1, err := app.worker("fakeSlave-1")
	if err != nil {
		t.Fatalf("unable to get worker, error %v", err)
	}
	if expected, got := NeedsRefresh, w1.state; expected == got {
		t.Fatalf("expected state does not match, expected %s got %s", expected, got)
	}
}

func TestPool_ItMarksWorkersToNotifyOnScaleDown(t *testing.T) {
	asg := &fakeAssigner{}
	call := &fakeCaller{err: errors.New("fake tiemout")}
	app := NewWorkerPool(asg, call)
	defer app.Terminate()

	app.UpdateSize(2)
	if !app.AddWorkerIfNotExists(0, "fakeSlave-0", net.ParseIP("127.0.0.1")) {
		t.Fatal("worker addition assertion expected true")
	}
	if !app.AddWorkerIfNotExists(1, "fakeSlave-1", net.ParseIP("127.0.0.2")) {
		t.Fatal("worker addition assertion expected true")
	}

	app.UpdateSize(1)
	app.RemoveWorkerByName("fakeSlave-1")

	if atLeast, got := int32(2), atomic.LoadInt32(&call.assigns); got < atLeast {
		t.Errorf("min Workloads calls not match, at least %d got %d", atLeast, got)
	}
	w0, err := app.worker("fakeSlave-0")
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

func (f *fakeCaller) RestartWorker(ctx context.Context, name string) error {
	return nil
}
