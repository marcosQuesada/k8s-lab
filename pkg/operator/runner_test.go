package operator

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestItConsumesProducedEntriesWithSuccess(t *testing.T) {
	var totalCalls int32
	f := func(context.Context, interface{}) error {
		atomic.AddInt32(&totalCalls, 1)
		return nil
	}
	r := NewRunner()
	r.workerFrequency = time.Millisecond * 50
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	go r.Run(ctx, f)
	r.Process("hello")
	time.Sleep(time.Millisecond * 200) // Let the worker run

	if expected, got := 1, atomic.LoadInt32(&totalCalls); expected != int(got) {
		t.Fatalf("unexpected totalCalls, expected %d gpt %d", expected, got)
	}
}

func TestItRetriesConsumedEntriesOnHandlingErrorUntilMaxRetries(t *testing.T) {
	var totalCalls int32
	f := func(context.Context, interface{}) error {
		atomic.AddInt32(&totalCalls, 1)
		return errors.New("foo error")
	}
	r := NewRunner()
	r.workerFrequency = time.Millisecond * 50
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go r.Run(ctx, f)
	r.Process("hello")
	time.Sleep(time.Millisecond * 200) // Let the worker run

	if expected, got := maxRetries, atomic.LoadInt32(&totalCalls); expected > int(got) {
		t.Fatalf("unexpected totalCalls, expected %d got %d", expected, got)
	}
}
