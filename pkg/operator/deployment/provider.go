package pod

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"time"
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

// Refresh updates deployment
func (p *Provider) Refresh(ctx context.Context, deploymentName string) error {
	data := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"%s"}}}}}`, time.Now().String())
	_, err := p.client.AppsV1().Deployments(p.namespace).Patch(ctx, deploymentName, types.StrategicMergePatchType, []byte(data), metav1.PatchOptions{FieldManager: "kubectl-rollout"})
	if err != nil {
		return fmt.Errorf("unable to patch deployment %s error %v", deploymentName, err)
	}

	return nil
}
