package config

import (
	"testing"
)

func TestWorkload_Equals(t *testing.T) {
	a0 := &Workload{Jobs: []Job{"foo", "bar", "zoom", "xxxx", "hmmmm"}}
	a1 := &Workload{Jobs: []Job{"bar", "foo", "xxxx", "hmmmm", "zoom"}}
	if !a0.Equals(a1) {
		t.Fatal("expected equal workloads")
	}
	if !a1.Equals(a0) {
		t.Fatal("expected equal workloads")
	}
}

func TestWorkload_Differences(t *testing.T) {
	a0 := &Workload{Jobs: []Job{"fss", "zoom", "xxxx", "hmmmm"}}
	a1 := &Workload{Jobs: []Job{"bar", "zzzz", "foo", "xxxx", "hmmmm", "zoom"}}

	i, e := a0.Difference(a1)
	if expected, got := 3, len(i); expected != got {
		t.Fatalf("includes size do not match, expected %d got %d", expected, got)
	}
	if expected, got := 1, len(e); expected != got {
		t.Fatalf("excludes size do not match, expected %d got %d", expected, got)
	}
	expectedIncluded := []Job{"bar", "zzzz", "foo"}
	if !jobsEquals(expectedIncluded, i) {
		t.Fatal("includes do not match")
	}
	expectedExcluded := []Job{"fss"}
	if !jobsEquals(expectedExcluded, e) {
		t.Fatal("excludes do not match")
	}
}

func TestWorkloads_ItFailsOnAssertingEqualsOnDifferentVersions(t *testing.T) {
	a0 := &Workloads{Version: 1}
	a1 := &Workloads{Version: 2}

	if a0.Equals(a1) {
		t.Error("not matching expectation")
	}
}

func TestWorkloads_Equals(t *testing.T) {
	ws0 := map[string]*Workload{
		"foo_0": &Workload{Jobs: []Job{"foo", "bar", "zoom", "xxxx", "hmmmm"}},
		"foo_1": &Workload{Jobs: []Job{"foo1", "bar1", "zoom1", "xxxx1", "hmmmm2"}},
	}
	ws1 := map[string]*Workload{
		"foo_0": &Workload{Jobs: []Job{"foo", "bar", "zoom", "xxxx", "hmmmm"}},
		"foo_1": &Workload{Jobs: []Job{"foo1", "bar1", "zoom1", "xxxx1", "hmmmm2"}},
	}

	w0 := &Workloads{
		Workloads: ws0,
		Version:   1,
	}
	w1 := &Workloads{
		Workloads: ws1,
		Version:   1,
	}
	if !w0.Equals(w1) {
		t.Fatal("expected equal workloads")
	}
	if !w1.Equals(w0) {
		t.Fatal("expected equal workloads")
	}
}
