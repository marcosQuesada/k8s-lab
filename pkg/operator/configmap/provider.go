package configmap

import (
	"bytes"
	"context"
	"fmt"
	cfg "github.com/marcosQuesada/k8s-lab/pkg/config"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const defaultConfigKey = "config.yml"

// Provider implements config repository in top of configmap
type Provider struct {
	client         kubernetes.Interface
	namespace      string
	configMapName  string
	deploymentName string
}

// NewProvider instantiate configmap provider
func NewProvider(cl kubernetes.Interface, namespace, configMapName, deploymentName string) *Provider {
	return &Provider{
		client:         cl,
		namespace:      namespace,
		configMapName:  configMapName,
		deploymentName: deploymentName,
	}
}

// Set updates workload assignation to configmap
func (p *Provider) Set(ctx context.Context, a *cfg.Workloads) error {
	cm, err := p.client.CoreV1().ConfigMaps(p.namespace).Get(ctx, p.configMapName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to get config map %v", err)
	}
	var buffer bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&buffer)
	yamlEncoder.SetIndent(2)
	if err := yamlEncoder.Encode(a); err != nil {
		return fmt.Errorf("unable to Marshall config map, error %v", err)
	}

	cm.Data[defaultConfigKey] = buffer.String()
	_, err = p.client.CoreV1().ConfigMaps(p.namespace).Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("unable to update config map %v", err)
	}

	return nil
}

// Get returns workload assignation from configmap
func (p *Provider) Get(ctx context.Context) (*cfg.Workloads, error) {
	cm, err := p.client.CoreV1().ConfigMaps(p.namespace).Get(ctx, p.configMapName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get config map %v", err)
	}
	w, err := p.decode(cm)
	if err != nil {
		return nil, fmt.Errorf("unable to decode workloads from config map %v", err)
	}
	context.Background()
	return w, nil
}

func (p *Provider) decode(cm *v1.ConfigMap) (*cfg.Workloads, error) {
	data := make(map[interface{}]interface{})
	if err := yaml.Unmarshal([]byte(cm.Data[defaultConfigKey]), &data); err != nil {
		return nil, fmt.Errorf("unable to unmarshall, error %v", err)
	}

	c := &cfg.Workloads{}
	if err := mapstructure.Decode(data, c); err != nil {
		return nil, fmt.Errorf("unable to decode, error %v", err)
	}

	return c, nil
}
