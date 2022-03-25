package crd

import (
	"context"
	"fmt"
	ap "github.com/marcosQuesada/k8s-lab/pkg/config"
	cfg "github.com/marcosQuesada/k8s-lab/pkg/config"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/apis/swarm/v1alpha1"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/generated/clientset/versioned"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type chainMember interface {
	Set(ctx context.Context, a *ap.Workloads) error
}

type ProviderMiddleware struct {
	prev, next chainMember
}

func NewProviderMiddleware(prev, next chainMember) *ProviderMiddleware {
	return &ProviderMiddleware{
		prev: prev,
		next: next,
	}
}

func (pm *ProviderMiddleware) Set(ctx context.Context, a *cfg.Workloads) error {
	if err := pm.prev.Set(ctx, a); err != nil {
		return err
	}
	if err := pm.next.Set(ctx, a); err != nil {
		return err
	}

	return nil
}

// Provider implements swarm access
type Provider struct {
	client    versioned.Interface
	namespace string
	swarmName string
}

// NewProvider instantiate configmap provider
func NewProvider(cl versioned.Interface, namespace, swarmName string) *Provider {
	return &Provider{
		client:    cl,
		namespace: namespace,
		swarmName: swarmName,
	}
}

// Set updates workload assignation to configmap
func (p *Provider) Set(ctx context.Context, a *cfg.Workloads) error {
	sw, err := p.client.K8slabV1alpha1().Swarms(p.namespace).Get(ctx, p.swarmName, metav1.GetOptions{})
	if err != nil {
		if _, ok := err.(*apiErrors.StatusError); !ok {
			return fmt.Errorf("unable to get swarm map %v", err)
		}

		sw, err = p.initializeCRD(ctx) // @TODO: REFACTOR
		if err != nil {
			return fmt.Errorf("unable to initialize CRD, error %v", err)
		}
	}

	sw.Spec.Version = a.Version
	sw.Spec.Size = len(a.Workloads)
	sw.Spec.Members = []v1alpha1.Worker{}
	for n, s := range a.Workloads {
		sw.Spec.Members = append(sw.Spec.Members, v1alpha1.Worker{
			Name:      n,
			Jobs:      p.adaptJobsToAlpha(s.Jobs),
			State:     v1alpha1.Status{Phase: v1alpha1.PhaseRunning},
			CreatedAt: time.Now().Unix(),
		})
	}

	// @TODO: Update as SubResource
	_, err = p.client.K8slabV1alpha1().Swarms(p.namespace).Update(ctx, sw, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("unable to update config map %v", err)
	}

	//swr.Status.Phase = "RUNNING"
	//_, err = p.client.K8slabV1alpha1().Swarms(p.namespace).UpdateStatus(ctx, swr, metav1.UpdateOptions{})
	//if err != nil {
	//	return fmt.Errorf("unable to update config map %v", err)
	//}

	return nil
}

// Get returns workload assignation from configmap
func (p *Provider) Get(ctx context.Context) (*cfg.Workloads, error) {
	sw, err := p.client.K8slabV1alpha1().Swarms(p.namespace).Get(ctx, p.swarmName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get config map %v", err)
	}
	w, err := p.decode(sw.Spec)
	if err != nil {
		return nil, fmt.Errorf("unable to decode workloads from config map %v", err)
	}
	context.Background()
	return w, nil
}

// @TODO: Refactor!
func (p *Provider) initializeCRD(ctx context.Context) (*v1alpha1.Swarm, error) {
	sw := &v1alpha1.Swarm{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Swarm",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.swarmName,
			Namespace: p.namespace,
		},
		Spec: v1alpha1.SwarmSpec{},
	}
	return p.client.K8slabV1alpha1().Swarms(p.namespace).Create(ctx, sw, metav1.CreateOptions{})
}

func (p *Provider) decode(sw v1alpha1.SwarmSpec) (*cfg.Workloads, error) {
	c := &cfg.Workloads{
		Workloads: make(map[string]*cfg.Workload),
		Version:   sw.Version,
	}
	for _, member := range sw.Members {
		c.Workloads[member.Name] = &cfg.Workload{
			Jobs: p.adaptJobsFromAlpha(member.Jobs),
		}
	}

	return c, nil
}

func (p *Provider) adaptJobsToAlpha(j []cfg.Job) (res []v1alpha1.Job) {
	for _, job := range j {
		res = append(res, v1alpha1.Job(job))
	}
	return
}

func (p *Provider) adaptJobsFromAlpha(j []v1alpha1.Job) (res []cfg.Job) {
	for _, job := range j {
		res = append(res, cfg.Job(job))
	}
	return
}
