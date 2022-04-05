package app

import (
	"context"
	"fmt"
	"github.com/marcosQuesada/k8s-lab/pkg/operator/pod"
	swapi "github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/apis/swarm/v1alpha1"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/generated/clientset/versioned"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/generated/listers/swarm/v1alpha1"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/statefulset"
	log "github.com/sirupsen/logrus"
	api "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appv1 "k8s.io/client-go/listers/apps/v1"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/util/workqueue"
	"time"
)

const maxRetries = 5

type Manager interface {
	Process(ctx context.Context, namespace, name string, version int64, workloads []swapi.Job)
	UpdateSize(ctx context.Context, namespace, name string, size int) (version int64, err error)
	Delete(ctx context.Context, namespace, name string)
}

// swarmController linearize incoming commands, concurrent processing wouldn't make sense
type swarmController struct {
	queue             workqueue.RateLimitingInterface
	swarmClient       versioned.Interface
	swarmLister       v1alpha1.SwarmLister
	statefulSetLister appv1.StatefulSetLister
	podLister         v1.PodLister
	selectorStore     statefulset.SelectorStore
	manager           Manager
}

func NewSwarmController(cl versioned.Interface, swl v1alpha1.SwarmLister, stsl appv1.StatefulSetLister, pl v1.PodLister, ss statefulset.SelectorStore, m Manager) *swarmController {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	return &swarmController{
		queue:             queue,
		swarmClient:       cl,
		swarmLister:       swl,
		statefulSetLister: stsl,
		podLister:         pl,
		selectorStore:     ss,
		manager:           m,
	}
}

// Process swarm entry happens on swarm creation or update
func (c *swarmController) Process(ctx context.Context, namespace, name string) error {
	c.queue.Add(newProcessEvent(namespace, name))
	return nil
}

// UpdatePoolSize happens on statefulSet size variation
func (c *swarmController) UpdatePoolSize(ctx context.Context, namespace, name string, size int) error {
	c.queue.Add(newUpdateEvent(namespace, name, size))
	return nil
}

// Delete happens on swarm deletion
func (c *swarmController) Delete(ctx context.Context, namespace, name string) error {
	c.queue.Add(newDeleteEvent(namespace, name))
	return nil
}

func (c *swarmController) Run(ctx context.Context) {
	defer c.queue.ShutDown()

	wait.UntilWithContext(ctx, c.worker, time.Second)
}

func (c *swarmController) worker(ctx context.Context) {
	for c.processNextItem(ctx) {
	}
}

func (c *swarmController) processNextItem(ctx context.Context) bool {
	e, quit := c.queue.Get()
	if quit {
		log.Error("Queue goes down!")
		return false
	}
	defer c.queue.Done(e)

	ev := e.(Event)
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err := c.handle(ctx, ev)
	if err == nil {
		c.queue.Forget(e)
		return true
	}

	if c.queue.NumRequeues(e) < maxRetries {
		log.Errorf("Error processing ev %s, retry. Error: %v", ev, err)
		c.queue.AddRateLimited(e)
		return true
	}

	log.Errorf("Error processing %s Max retries achieved: %v", ev, err)
	c.queue.Forget(e)
	utilruntime.HandleError(err)

	return true
}

func (c *swarmController) handle(ctx context.Context, ev Event) error {
	switch e := ev.(type) {
	case processEvent:
		return c.process(ctx, e.namespace, e.name)
	case updateEvent:
		return c.updatePool(ctx, e.namespace, e.name, e.size)
	case deleteEvent:
		return c.delete(ctx, e.namespace, e.name)
	}

	return fmt.Errorf("action %T not handled", ev)
}

func (c *swarmController) process(ctx context.Context, namespace, name string) error {
	log.Infof("Process swarm %s %s", namespace, name)
	sw, err := c.swarmLister.Swarms(namespace).Get(name)
	if err != nil {
		return fmt.Errorf("unable to get swarm namespace %s name %s", namespace, name)
	}

	log.Infof("Processing swarm %s namespace %s statefulset name %s configmap name %s version %d total workloads %d",
		sw.Name, sw.Namespace, sw.Spec.StatefulSetName, sw.Spec.ConfigMapName, sw.Spec.Version, len(sw.Spec.Workload))

	sts, err := c.statefulSetLister.StatefulSets(sw.Namespace).Get(sw.Spec.StatefulSetName)
	if err != nil {
		return fmt.Errorf("unable to get statefulset from swarm %s on namespace %s error %v", name, namespace, err)
	}

	if err := c.selectorStore.Register(namespace, sts.Name, sts.Spec.Selector); err != nil {
		return fmt.Errorf("unable to register key %s %s error %v", namespace, sts.Name, err)
	}

	names, err := c.podNamesFromSelector(namespace, sts.Spec.Selector)
	if err != nil {
		return fmt.Errorf("unable to get pods from selector, error %v", err)
	}

	log.Infof("Controller found size %d worker pods %s", len(names), names)

	c.manager.Process(ctx, namespace, name, sw.Spec.Version, sw.Spec.Workload)

	return c.updatePool(ctx, namespace, sts.Name, int(*sts.Spec.Replicas))
}

func (c *swarmController) updatePool(ctx context.Context, namespace, name string, size int) error {
	log.Infof("Update swarm %s %s", namespace, name)

	swarmName, err := c.swarmNameFromStatefulSetName(namespace, name)
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
	sw, err := c.swarmLister.Swarms(namespace).Get(name)
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

func (c *swarmController) podNamesFromSelector(namespace string, ls *metav1.LabelSelector) ([]string, error) {
	selector, err := metav1.LabelSelectorAsSelector(ls)
	if err != nil {
		return nil, fmt.Errorf("unable to get label selector, error %v", err)
	}

	pods, err := c.podLister.Pods(namespace).List(selector)
	if err != nil {
		return nil, fmt.Errorf("unable to get pods from selector, error %v", err)
	}

	names := []string{}
	for _, pd := range pods {
		if pod.IsTerminated(pd) || pod.HasDeletionTimestamp(pd) {
			continue
		}
		names = append(names, pd.Name)
	}

	return names, nil
}

func (c *swarmController) statefulSetFromSwarmName(namespace, name string) (*api.StatefulSet, error) {
	sw, err := c.swarmLister.Swarms(namespace).Get(name)
	if err != nil {
		return nil, fmt.Errorf("unable to get swarm namespace %s name %s", namespace, name)
	}

	log.Infof("Processing swarm %s namespace %s statefulset name %s configmap name %s version %d workloads %v",
		sw.Name, sw.Namespace, sw.Spec.StatefulSetName, sw.Spec.ConfigMapName, sw.Spec.Version, sw.Spec.Workload)

	sts, err := c.statefulSetLister.StatefulSets(sw.Namespace).Get(sw.Spec.StatefulSetName)
	if err != nil {
		return nil, fmt.Errorf("unable to get statefulset from swarm %s on namespace %s name %s error %v", sw.Name, sw.Namespace, sw.Spec.StatefulSetName, err)
	}
	return sts, nil
}

func (c *swarmController) swarmNameFromStatefulSetName(namespace, name string) (string, error) {
	sws, err := c.swarmLister.Swarms(namespace).List(labels.NewSelector())
	if err != nil {
		return "", fmt.Errorf("unable to list swarm namespace %s name %s", namespace, name)
	}

	for _, sw := range sws {
		if sw.Spec.StatefulSetName == name {
			return sw.Name, nil
		}
	}

	return "", fmt.Errorf("unable to find swarm name from statefulset %s", name)
}
