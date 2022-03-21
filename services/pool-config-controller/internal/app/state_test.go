package app

import (
	config2 "github.com/marcosQuesada/k8s-lab/pkg/config"
	"testing"
)

const fakeWorkerName = "foo"

var jobs = []config2.Job{
	"stream:xxrtve1",
	"stream:xxrtve2",
	"stream:zrtve2",
	"stream:zrtve1",
	"stream:zrtve0",
	"stream:cctv0",
	"stream:cctv1",
	"stream:cctv2",
	"stream:xxctv3",
	"stream:cctv3",
	"stream:xxctv0",
	"stream:xxctv10",
	"stream:xxctv11",
	"stream:xxctv12",
	"stream:xxctv13",
	"stream:xxctv14",
	"stream:yxctv0",
	"stream:yxctv1",
	"stream:yxctv2",
	"stream:yxctv3",
	"stream:xabcn0",
	"stream:xacb01",
	"stream:xacb02",
	"stream:xacb03",
	"stream:xacb04",
	"stream:sportnews0",
	"stream:sportnews1",
}

func TestDiffCalculationOnStaticSet(t *testing.T) {
	set := []config2.Job{
		"discussions",
		"stream:sportnews1",
		"comments",
		"aatoupdate_temp",
		"creditnotes",
		"boards",
		"countrytaxes",
		"stream:xxctv12",
		"stream:yxctv3",
		"recurringstream:xxctv11",
		"stream:xacb01",
		"mailinglog",
		"stream:xxctv11",
		"stream:yxctv2",
		"orders",
	}

	app := NewState(set, fakeWorkerName)
	totalWorkers := 3
	var version int64 = 1
	if err := app.BalanceWorkload(totalWorkers, version); err != nil {
		t.Fatalf("unable to balance keys %v", err)
	}

	for i := 0; i < totalWorkers; i++ {
		asg, err := app.Workload(i)
		if err != nil {
			t.Fatalf("worker %d not found", i)
		}

		if expected, got := len(set)/3, len(asg.Jobs); expected != got {
			t.Fatalf("unexpected total entries on subset 0, expected %d got %d", expected, got)
		}
	}
}

func TestDiffCalculationOnNounStaticSet(t *testing.T) {
	set := []config2.Job{
		"discussions",
		"stream:sportnews1",
		"comments",
		"aatoupdate_temp",
		"creditnotes",
		"boards",
		"countrytaxes",
		"stream:xxctv12",
		"stream:yxctv3",
	}
	app := NewState(set, fakeWorkerName)
	totalWorkers := 2
	var version int64 = 1
	if err := app.BalanceWorkload(totalWorkers, version); err != nil {
		t.Fatalf("unable to balance keys %v", err)
	}

	asg, err := app.Workload(0)
	if err != nil {
		t.Fatalf("unable to complete balance %v", err)
	}

	if expected, got := 5, len(asg.Jobs); expected != got {
		t.Fatalf("unexpected total entries on subset 0, expected %d got %d", expected, got)
	}

	asg, err = app.Workload(1)
	if err != nil {
		t.Fatalf("unable to complete balance %v", err)
	}

	if expected, got := 4, len(asg.Jobs); expected != got {
		t.Fatalf("unexpected total entries on subset 0, expected %d got %d", expected, got)
	}
}

func TestShardingByTotalWorkersOnASingleNodeComposition(t *testing.T) {
	app := NewState(jobs, fakeWorkerName)
	totalWorkers := 1
	var version int64 = 1
	if err := app.BalanceWorkload(totalWorkers, version); err != nil {
		t.Fatalf("unable to balance keys %v", err)
	}

	asg, err := app.Workload(0)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if includes, got := len(jobs), len(asg.Jobs); got != includes {
		t.Fatalf("unexpected total excludes, expected %d git %d", includes, got)
	}
}

func TestOnScalingUpWorkersOnceAssigned(t *testing.T) {
	app := NewState(jobs, fakeWorkerName)
	totalWorkers := 2
	var version int64 = 1
	if err := app.BalanceWorkload(totalWorkers, version); err != nil {
		t.Fatalf("unable to balance keys %v", err)
	}
	a0, err := app.Workload(0)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if expected, got := 14, len(a0.Jobs); got != expected {
		t.Errorf("Includes do not match, expected %d got %d", expected, got)
	}

	a1, err := app.Workload(1)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if expected, got := 13, len(a1.Jobs); got != expected {
		t.Errorf("Includes do not match, expected %d got %d", expected, got)
	}

	totalWorkers = 3
	version = 2
	if err := app.BalanceWorkload(totalWorkers, version); err != nil {
		t.Fatalf("unable to balance keys %v", err)
	}

	a2, err := app.Workload(2)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if expected, got := 9, len(a2.Jobs); got != expected {
		t.Errorf("Includes do not match, expected %d got %d", expected, got)
	}
}

/**
3 Nodes : 9 9 9
2 Nodes:  14, 13
- 0: inc: 14    exc: 0
- 1: inc: 13    exc: 5
*/
func TestOnScalingDownWorkersOnceAssigned(t *testing.T) {
	app := NewState(jobs, fakeWorkerName)
	totalWorkers := 3
	var version int64 = 1
	if err := app.BalanceWorkload(totalWorkers, version); err != nil {
		t.Fatalf("unable to balance keys %v", err)
	}

	totalWorkers = 2
	version = 2
	if err := app.BalanceWorkload(totalWorkers, version); err != nil {
		t.Fatalf("unable to balance keys %v", err)
	}

	a0, err := app.Workload(0)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if expected, got := 14, len(a0.Jobs); got != expected {
		t.Errorf("Includes do not match, expected %d got %d", expected, got)
	}

	a1, err := app.Workload(1)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if expected, got := 13, len(a1.Jobs); got != expected {
		t.Errorf("Includes do not match, expected %d got %d", expected, got)
	}

	_, err = app.Workload(2)
	if err == nil {
		t.Fatal("expected error not found")
	}
}

func TestOnScalingToZeroWorkersOnceAssigned(t *testing.T) {
	app := NewState(jobs, fakeWorkerName)
	totalWorkers := 3
	var version int64 = 1
	if err := app.BalanceWorkload(totalWorkers, version); err != nil {
		t.Fatalf("unable to balance keys %v", err)
	}

	totalWorkers = 0
	version = 2
	if err := app.BalanceWorkload(totalWorkers, version); err != nil {
		t.Fatalf("unable to balance keys %v", err)
	}

	if expected, got := 0, app.size(); expected != got {
		t.Errorf("size does not match, expected %d got %d", expected, got)
	}
}

var realScenarioBug = []config2.Job{
	"stream:xxrtve1",
	"stream:xxrtve2",
	"stream:zrtve2:new",
	"stream:zrtve1:new",
	"stream:zrtve0:new",
	"stream:cctv0:updated",
	"stream:history:new",
	"stream:foo:new",
	"stream:xxctv3:new",
	"stream:cctv3:updated",
	"stream:xxctv0:updated",
	"stream:xxctv10:updated",
	"stream:xxctv11:updated",
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
	"stream:cars:updated",
}

func TestScalingUpOnNounWorkloadSizeAssignsExpectedJobs(t *testing.T) {
	app := NewState(realScenarioBug, fakeWorkerName)
	totalWorkers := 2
	var version int64 = 1
	if err := app.BalanceWorkload(totalWorkers, version); err != nil {
		t.Fatalf("unable to balance keys %v", err)
	}

	totalWorkers = 12
	version = 2
	if err := app.BalanceWorkload(totalWorkers, version); err != nil {
		t.Fatalf("unable to balance keys %v", err)
	}
	a0, err := app.Workload(0)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if expected, got := 3, len(a0.Jobs); got != expected {
		t.Errorf("Includes do not match, expected %d got %d", expected, got)
	}
	a3, err := app.Workload(3)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if expected, got := 2, len(a3.Jobs); got != expected {
		t.Errorf("Includes do not match, expected %d got %d", expected, got)
	}
}
