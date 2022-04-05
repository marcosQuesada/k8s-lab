package statefulset

import (
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestSelectorStore_ItRegistersNewSelectorOnStore(t *testing.T) {
	sl := fakeSelector("app", "foo")
	namespace := "default"
	name := "foo-workers"

	ss := NewSelectorStore().(*selectorStore)
	if err := ss.Register(namespace, name, sl); err != nil {
		t.Fatalf("unable to register, error %v", err)
	}

	if expected, got := 1, ss.len(); expected != got {
		t.Fatalf("store size does not match, expected %d got %d", expected, got)
	}
}

func TestSelectorStore_ItMarchesLabelFromRegisteredSelector(t *testing.T) {
	key, value := "app", "foo"
	sl := fakeSelector(key, value)
	namespace := "default"
	name := "foo-workers"

	ss := NewSelectorStore().(*selectorStore)
	if err := ss.Register(namespace, name, sl); err != nil {
		t.Fatalf("unable to register, error %v", err)
	}

	p := getFakePod(namespace, "foo-worker", map[string]string{key: value})
	if !ss.Matches(namespace, name, p.Labels) {
		t.Fatal("expected selector match")
	}
}

func TestSelectorStore_ItUnRegistersSelectorFromStore(t *testing.T) {
	sl := fakeSelector("app", "foo")
	namespace := "default"
	name := "foo-workers"

	ss := NewSelectorStore().(*selectorStore)
	if err := ss.Register(namespace, name, sl); err != nil {
		t.Fatalf("unable to register, error %v", err)
	}

	ss.UnRegister(namespace, name)
	if expected, got := 0, ss.len(); expected != got {
		t.Fatalf("store size does not match, expected %d got %d", expected, got)
	}
}

func fakeSelector(key, value string) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{key: value},
	}
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
