package pod

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Provider deletes pod to force pod fresh recreation
type Provider struct {
	client    kubernetes.Interface
	namespace string
}

// NewProvider instantiates pod refresher provider
func NewProvider(cl kubernetes.Interface, namespace string) *Provider {
	return &Provider{
		client:    cl,
		namespace: namespace,
	}
}

// RefreshPod deletes pod
func (p *Provider) RefreshPod(ctx context.Context, name string) error {
	err := p.client.CoreV1().Pods(p.namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("unable to delete pod %s error %v", name, err)
	}

	return nil
}
