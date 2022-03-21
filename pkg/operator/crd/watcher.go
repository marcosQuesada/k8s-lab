package crd

import (
	"context"
	"github.com/marcosQuesada/k8s-lab/pkg/operator"
	"github.com/marcosQuesada/k8s-lab/services/pool-config-controller/internal/infra/k8s/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

type listWatcherAdapter struct {
	client    versioned.Interface
	namespace string
}

// NewListWatcherAdapter instantiates swarm list watcher adapter
func NewListWatcherAdapter(c versioned.Interface, namespace string) operator.ListWatcher {
	return &listWatcherAdapter{
		client:    c,
		namespace: namespace,
	}
}

// List handles swarm listing in the specified namespace
func (a *listWatcherAdapter) List(options metav1.ListOptions) (runtime.Object, error) {
	return a.client.K8slabV1alpha1().Swarms(a.namespace).List(context.Background(), options)
}

// Watch creates a stream of swarm events in watched namespace
func (a *listWatcherAdapter) Watch(options metav1.ListOptions) (watch.Interface, error) {
	return a.client.K8slabV1alpha1().Swarms(a.namespace).Watch(context.Background(), options)
}
