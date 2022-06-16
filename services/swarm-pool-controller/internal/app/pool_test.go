package app

import (
	"context"
	"github.com/marcosQuesada/k8s-lab/pkg/config"
	"sync"
	"sync/atomic"
	"testing"
)

var version int64 = 1

func TestOnWorkerPoolUpdateStateGetsBalancedAndVersionUpdated(t *testing.T) {
	asg := &fakeAssigner{}
	call := &fakeCaller{}
	p := newWorkerPool(version, asg, call)

	size := 2
	v, err := p.UpdateSize(context.Background(), size)
	if err != nil {
		t.Fatalf("unable to process size, error %v", err)
	}

	newVersion := 2
	if expected, got := newVersion, int(v); expected != got {
		t.Errorf("version does not match, expected %d got %d", expected, got)
	}

	if expected, got := size, p.Size(); expected != got {
		t.Errorf("size does not match, expected %d got %d", expected, got)
	}
}

type fakeAssigner struct {
	balanceRequests int32
	workloads       *config.Workloads
}

func (a *fakeAssigner) BalanceWorkload(totalWorkers int, version int64) (*config.Workloads, error) {
	atomic.AddInt32(&a.balanceRequests, 1)
	return a.workloads, nil
}

func (a *fakeAssigner) Workloads() *config.Workloads {
	return &config.Workloads{
		Workloads: map[string]*config.Workload{"fake_0": &config.Workload{}},
		Version:   1,
	}
}

type fakeCaller struct {
	assigns     int32
	err         error
	assignation *config.Workloads
	mutex       sync.RWMutex
}

func (f *fakeCaller) Assign(ctx context.Context, namespace, name string, w *config.Workloads) (err error) {
	atomic.AddInt32(&f.assigns, 1)
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.assignation = w

	return nil
}

func (f *fakeCaller) RestartWorker(ctx context.Context, namespace, name string) error {
	return nil
}
