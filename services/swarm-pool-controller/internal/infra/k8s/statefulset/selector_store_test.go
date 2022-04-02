package statefulset

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"testing"
)

func TestSelectorStore_ItRegistersNewSelectorOnStore(t *testing.T) {
	sl := fakeSelector("app", "foo")
	namespace := "default"
	name := "foo-workers"
	swarmName := "foo-swarm"

	ss := NewSelectorStore().(*selectorStore)
	ss.EnsureRegister(namespace, name, sl, swarmName)

	if expected, got := 1, ss.len(); expected != got {
		t.Fatalf("store size does not match, expected %d got %d", expected, got)
	}
}

func TestSelectorStore_ItMarchesLabelFromRegisteredSelector(t *testing.T) {
	key, value := "app", "foo"
	sl := fakeSelector(key, value)
	namespace := "default"
	name := "foo-workers"
	swarmName := "foo-swarm"

	ss := NewSelectorStore().(*selectorStore)
	ss.EnsureRegister(namespace, name, sl, swarmName)

	p := getFakePod(namespace, "foo-worker", map[string]string{key: value})
	if !ss.Matches(namespace, name, p.Labels) {
		t.Fatal("expected selector match")
	}
}

func TestSelectorStore_ItUnRegistersSelectorFromStore(t *testing.T) {
	sl := fakeSelector("app", "foo")
	namespace := "default"
	name := "foo-workers"
	swarmName := "foo-swarm"

	ss := NewSelectorStore().(*selectorStore)
	ss.EnsureRegister(namespace, name, sl, swarmName)

	ss.UnRegister(namespace, name)
	if expected, got := 0, ss.len(); expected != got {
		t.Fatalf("store size does not match, expected %d got %d", expected, got)
	}
}

func fakeSelector(key, value string) labels.Selector {
	l, _ := labels.NewRequirement(key, selection.In, []string{value})
	sl := labels.NewSelector()
	sl.Add(*l)
	return sl
}

func getFakePod(namespace, name string, labels map[string]string) apiv1.Pod {
	return apiv1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Name:            "nginx",
					Image:           "nginx",
					ImagePullPolicy: "Always",
				},
			},
			RestartPolicy: apiv1.RestartPolicyNever,
		},
	}
}
