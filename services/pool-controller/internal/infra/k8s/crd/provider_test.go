package crd

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/marcosQuesada/k8s-lab/pkg/config"
	"github.com/marcosQuesada/k8s-lab/services/pool-controller/internal/infra/k8s"
	"github.com/marcosQuesada/k8s-lab/services/pool-controller/internal/infra/k8s/apis/swarm/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestNewProvider_ItUpdatesConfigMapOnAssignWorkload(t *testing.T) {
	var namespace = "swarm"
	var name = "swarm-worker"

	clientset := k8s.BuildSwarmExternalClient()
	p := NewProvider(clientset, namespace, name)

	w := &config.Workloads{
		Version: 1,
		Workloads: map[string]*config.Workload{
			"swarm-worker-0": {Jobs: []config.Job{
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
			}},
			"swarm-worker-1": {Jobs: []config.Job{
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
			}},
		},
	}

	if err := p.Set(context.Background(), w); err != nil {
		t.Fatalf("unexepcted error setting workload %v, got %v", w, err)
	}
}

func TestNewProvider_ItGetsCRDWorkload(t *testing.T) {
	var namespace = "swarm"
	var name = "swarm-worker"

	clientset := k8s.BuildSwarmExternalClient()
	p := NewProvider(clientset, namespace, name)

	sw, err := p.Get(context.Background())
	if err != nil {
		t.Fatalf("unexepcted error setting workload %v, got %v", sw, err)
	}

	spew.Dump(sw)
}

func TestNewProvider_ItCratesSwarm(t *testing.T) {
	var namespace = "swarm"
	var name = "foo-bar-worker"

	clientset := k8s.BuildSwarmExternalClient()

	sw := &v1alpha1.Swarm{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Swarm",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.SwarmSpec{
			Version:      1,
			ExpectedSize: 3,
			Size:         2,
			Members: []v1alpha1.Member{{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Name:       "foobar_0",
				Jobs: []v1alpha1.Job{
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
				},
				State:     v1alpha1.MemberStatus{},
				CreatedAt: time.Now().Unix(),
			},
			},
		},
	}
	sw, err := clientset.K8slabV1alpha1().Swarms(namespace).Create(context.Background(), sw, metav1.CreateOptions{})

	clientset.K8slabV1alpha1().Swarms(namespace)

	w, err := clientset.K8slabV1alpha1().Swarms(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("unexepcted error setting workload %v, got %v", sw, err)
	}

	spew.Dump(w)
}
