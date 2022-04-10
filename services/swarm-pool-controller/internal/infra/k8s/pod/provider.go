package pod

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/k8s-lab/pkg/config"
	"github.com/marcosQuesada/k8s-lab/pkg/operator/pod"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

const catConfigCommand = "cat /app/config/config.yml"

// Provider deletes pod to force pod fresh recreation
type Provider struct {
	client kubernetes.Interface
	config *restclient.Config
}

// NewProvider instantiates pod refresher provider
func NewProvider(cl kubernetes.Interface, config *restclient.Config) *Provider {
	return &Provider{
		client: cl,
		config: config,
	}
}

// Refresh deletes pod to force restart on the latest version
func (p *Provider) Refresh(ctx context.Context, namespace, name string) error {
	err := p.client.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("unable to delete pod %s error %v", name, err)
	}

	return nil
}

// Workload gets pod config workload
func (p *Provider) Workload(ctx context.Context, namespace, name string) (*config.Workloads, error) {
	r, err := pod.ExecCmd(p.client, p.config, namespace, name, catConfigCommand)
	if err != nil {
		return nil, errors.Wrap(err, "unable to execute command")
	}

	w, err := config.Decode(r)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode")
	}

	return w, nil
}
