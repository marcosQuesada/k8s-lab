package operator

import (
	"sync/atomic"
	"testing"
)

func TestIfForwardsAddRequestWithProperLabels(t *testing.T) {
	label := "foo"
	feh := &fakeEventHandler{}
	s := NewLabelSelectorMiddleware(label, feh)

	p := getFakePod("default", "fake")
	p.ObjectMeta.Labels = map[string]string{"app": label}
	s.Add(&p)

	if expected, got := 1, feh.created(); expected != int(got) {
		t.Errorf("total does not match, expected %d got %d", expected, got)
	}
}
func TestItSkipsAdditionRequestWithEmptyLabels(t *testing.T) {
	label := "foo"
	feh := &fakeEventHandler{}
	s := NewLabelSelectorMiddleware(label, feh)

	p := getFakePod("default", "fake")
	s.Add(&p)

	if expected, got := 0, feh.created(); expected != int(got) {
		t.Errorf("total does not match, expected %d got %d", expected, got)
	}
}

type fakeEventHandler struct {
	totalCreated int32
	totalUpdated int32
	totalDeleted int32
}

func (f *fakeEventHandler) Add(obj interface{}) {
	atomic.AddInt32(&f.totalCreated, 1)
}

func (f *fakeEventHandler) Update(oldObj, newObj interface{}) {
	atomic.AddInt32(&f.totalUpdated, 1)
}

func (f *fakeEventHandler) Delete(obj interface{}) {
	atomic.AddInt32(&f.totalDeleted, 1)
}

func (f *fakeEventHandler) created() int32 {
	return atomic.LoadInt32(&f.totalCreated)
}
func (f *fakeEventHandler) updated() int32 {
	return atomic.LoadInt32(&f.totalUpdated)
}
func (f *fakeEventHandler) deleted() int32 {
	return atomic.LoadInt32(&f.totalDeleted)
}
