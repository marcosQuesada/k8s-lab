package app

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/k8s-lab/pkg/operator"
	swapi "github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/apis/swarm/v1alpha1"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/generated/clientset/versioned"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/statefulset"
	log "github.com/sirupsen/logrus"
	api "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Manager interface {
	Process(ctx context.Context, namespace, name string, version int64, workloads []swapi.Job)
	UpdateSize(ctx context.Context, namespace, name string, size int) (version int64, err error)
	Delete(ctx context.Context, namespace, name string)
}

type Provider interface {
	Swarm(namespace, name string) (*swapi.Swarm, error)
	StatefulSet(namespace, name string) (*api.StatefulSet, error)
	PodNamesFromSelector(namespace string, ls *metav1.LabelSelector) ([]string, error)
	SwarmNameFromStatefulSetName(namespace, name string) (string, error)
}

// swarmController linearize incoming commands, concurrent processing wouldn't make sense
type swarmController struct {
	swarmClient   versioned.Interface
	selectorStore statefulset.SelectorStore
	manager       Manager
	provider      Provider
	runner        operator.Runner
}

func NewSwarmController(cl versioned.Interface, ss statefulset.SelectorStore, m Manager, p Provider, r operator.Runner) *swarmController {
	return &swarmController{
		swarmClient:   cl,
		selectorStore: ss,
		manager:       m,
		provider:      p,
		runner:        r,
	}
}

// Create swarm entry happens on swarm creation
func (c *swarmController) Create(ctx context.Context, namespace, name string) error {
	c.runner.Process(newProcessSwarm(namespace, name))
	return nil
}

// Update swarm entry happens on swarm update
func (c *swarmController) Update(ctx context.Context, namespace, name string) error {
	c.runner.Process(newProcessSwarm(namespace, name)) // @TODO: Improve this!
	return nil
}

// UpdatePoolSize happens on statefulSet size variation
func (c *swarmController) UpdatePoolSize(ctx context.Context, namespace, name string, size int) error {
	c.runner.Process(newUpdateSwarmSize(namespace, name, size))
	return nil
}

// Delete happens on swarm deletion
func (c *swarmController) Delete(ctx context.Context, namespace, name string) error {
	c.runner.Process(newDeleteSwarm(namespace, name))
	return nil
}

func (c *swarmController) Run(ctx context.Context) {
	c.runner.Run(ctx, c.handle)
}

func (c *swarmController) handle(ctx context.Context, e interface{}) error {
	ev := e.(Event)
	switch e := ev.(type) {
	case processSwarm:
		return c.process(ctx, e.namespace, e.name)
	case updateSwarmSize:
		return c.updatePool(ctx, e.namespace, e.name, e.size)
	case deleteSwarm:
		return c.delete(ctx, e.namespace, e.name)
	}

	return fmt.Errorf("action %T not handled", ev)
}

// process happens on swarm create/update event
func (c *swarmController) process(ctx context.Context, namespace, name string) error {
	log.Infof("Create swarm %s %s", namespace, name)
	sw, err := c.provider.Swarm(namespace, name)
	if err != nil {
		return err
	}

	log.Infof("Processing swarm %s namespace %s statefulset name %s configmap name %s version %d total workloads %d",
		sw.Name, sw.Namespace, sw.Spec.StatefulSetName, sw.Spec.ConfigMapName, sw.Spec.Version, len(sw.Spec.Workload))

	sts, err := c.provider.StatefulSet(sw.Namespace, sw.Spec.StatefulSetName)
	if err != nil {
		return fmt.Errorf("unable to get statefulset from swarm %s on namespace %s error %v", name, namespace, err)
	}

	// idempotent call
	if err := c.selectorStore.Register(namespace, sts.Name, sts.Spec.Selector); err != nil {
		return fmt.Errorf("unable to register key %s %s error %v", namespace, sts.Name, err)
	}

	names, err := c.provider.PodNamesFromSelector(namespace, sts.Spec.Selector)
	if err != nil {
		return fmt.Errorf("unable to get pods from selector, error %v", err)
	}

	log.Infof("Controller found size %d worker pods %s", len(names), names)

	c.manager.Process(ctx, namespace, name, sw.Spec.Version, sw.Spec.Workload)

	return c.updatePool(ctx, namespace, sts.Name, int(*sts.Spec.Replicas))
}

func (c *swarmController) updatePool(ctx context.Context, namespace, name string, size int) error {
	log.Infof("Update swarm %s %s", namespace, name)

	swarmName, err := c.provider.SwarmNameFromStatefulSetName(namespace, name)
	if err != nil {
		return fmt.Errorf("unable to get swarm error %v", err)
	}

	version, err := c.manager.UpdateSize(ctx, namespace, swarmName, size)
	if err != nil {
		return fmt.Errorf("unable to update swarm %s size error %v", swarmName, err)
	}

	_, err = c.updateSwarm(ctx, namespace, swarmName, version, size)
	if err != nil {
		return fmt.Errorf("unable to update swarm %s error %v", name, err)
	}
	return nil
}

func (c *swarmController) delete(ctx context.Context, namespace, name string) error {
	c.selectorStore.UnRegister(namespace, name)
	c.manager.Delete(ctx, namespace, name)
	return nil
}

func (c *swarmController) updateSwarm(ctx context.Context, namespace, name string, version int64, size int) (*swapi.Swarm, error) {
	sw, err := c.provider.Swarm(namespace, name)
	if err != nil {
		return nil, fmt.Errorf("unable to get swarm namespace %s name %s", namespace, name)
	}

	log.Infof("Update swarm %s namespace %s statefulset name %s configmap name %s version %d workloads %d Phase %s",
		name, namespace, sw.Spec.StatefulSetName, sw.Spec.ConfigMapName, sw.Spec.Version, len(sw.Spec.Workload), sw.Status.Phase)

	swr := sw.DeepCopyObject()
	updatedSwarm := swr.(*swapi.Swarm)
	updatedSwarm.Spec.Size = size
	updatedSwarm.Spec.Version = version
	swu, err := c.swarmClient.K8slabV1alpha1().Swarms(namespace).Update(ctx, updatedSwarm, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to update swarm %s error %v", sw.Name, err)
	}

	return swu, nil
}
