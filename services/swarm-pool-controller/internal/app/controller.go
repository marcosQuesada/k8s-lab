package app

import (
	"context"
	"fmt"
	swapi "github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/apis/swarm/v1alpha1"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/generated/clientset/versioned"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/crd/generated/listers/swarm/v1alpha1"
	"github.com/marcosQuesada/k8s-lab/services/swarm-pool-controller/internal/infra/k8s/statefulset"
	log "github.com/sirupsen/logrus"
	api "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appv1 "k8s.io/client-go/listers/apps/v1"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/util/workqueue"
	"time"
)

const maxRetries = 5

type action string

const processSwarm = action("processSwarm")
const updatePool = action("updatePool")

type event struct {
	action    action
	namespace string
	name      string
}

type swarmController struct {
	queue             workqueue.RateLimitingInterface
	swarmClient       versioned.Interface
	swarmLister       v1alpha1.SwarmLister
	statefulSetLister appv1.StatefulSetLister
	podLister         v1.PodLister
	selectorStore     statefulset.SelectorStore
}

func NewSwarmController(cl versioned.Interface, swl v1alpha1.SwarmLister, stsl appv1.StatefulSetLister, pl v1.PodLister, ss statefulset.SelectorStore) *swarmController {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	return &swarmController{
		queue:             queue,
		swarmClient:       cl,
		swarmLister:       swl,
		statefulSetLister: stsl,
		podLister:         pl,
		selectorStore:     ss,
	}
}

// Process swarm entry happens on swarm creation @TODO: Update?
func (c *swarmController) Process(ctx context.Context, namespace, name string) error {
	c.enqueue(processSwarm, namespace, name)
	return nil
}

func (c *swarmController) UpdatePool(ctx context.Context, namespace, name string) error {
	c.enqueue(updatePool, namespace, name)
	return nil
}

func (c *swarmController) process(ctx context.Context, namespace, name string) error {
	sw, err := c.swarmLister.Swarms(namespace).Get(name)
	if err != nil {
		return fmt.Errorf("unable to get swarm namespace %s name %s", namespace, name)
	}

	log.Infof("Processing swarm %s namespace %s statefulset name %s configmap name %s version %d workloads %d Phase %s",
		sw.Name, sw.Spec.Namespace, sw.Spec.StatefulSetName, sw.Spec.ConfigMapName, sw.Spec.Version, len(sw.Spec.Workload), sw.Status.Phase)

	sts, err := c.statefulSetLister.StatefulSets(sw.Spec.Namespace).Get(sw.Spec.StatefulSetName)
	if err != nil {
		return fmt.Errorf("unable to get statefulset from swarm %s on namespace %s name %s error %v", sw.Name, sw.Spec.Namespace, sw.Spec.StatefulSetName, err)
	}

	selector, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	if err != nil {
		return fmt.Errorf("unable to get label selector, error %v", err)
	}

	c.selectorStore.EnsureRegister(sts.Namespace, sts.Name, selector, name)

	if sw.Spec.Size == int(*sts.Spec.Replicas) && sw.Status.Phase == swapi.PhaseRunning {
		return nil
	}

	pods, err := c.podLister.Pods(sts.Namespace).List(selector)
	if err != nil {
		return fmt.Errorf("unable to get pods from selector, error %v", err)
	}

	names := []string{}
	for _, pod := range pods {
		names = append(names, pod.Name)
	}
	log.Infof("Controller found size %d worker pods %s", len(names), names) // @TODO: Needs to get state before computing

	swr := sw.DeepCopyObject()
	updatedSwarm := swr.(*swapi.Swarm)
	updatedSwarm.Spec.Size = int(*sts.Spec.Replicas)
	upsw, err := c.swarmClient.K8slabV1alpha1().Swarms(namespace).Update(ctx, updatedSwarm, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("unable to update swarm %s error %v", sw.Name, err)
	}

	log.Infof("Process swarm updated %s", name)

	upswr := upsw.DeepCopyObject()
	upSwarm := upswr.(*swapi.Swarm)
	upSwarm.Status.Phase = swapi.PhaseRunning
	_, err = c.swarmClient.K8slabV1alpha1().Swarms(namespace).UpdateStatus(ctx, upSwarm, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("unable to update swarm status map %v", err)
	}
	log.Infof("Process swarm status updated  %s", name)
	return nil
}

func (c *swarmController) Matches(namespace, name string, l map[string]string) bool {
	return c.selectorStore.Matches(namespace, name, l)
}

func (c *swarmController) updatePool(ctx context.Context, namespace, name string) error {
	swarmName, ok := c.selectorStore.SwarmName(namespace, name)
	if !ok {
		return nil
	}

	sw, err := c.swarmLister.Swarms(namespace).Get(swarmName)
	if err != nil {
		return fmt.Errorf("unable to get swarm namespace %s name %s", namespace, name)
	}

	swr := sw.DeepCopyObject()
	updatedSwarm := swr.(*swapi.Swarm)
	updatedSwarm.Status.Phase = swapi.PhaseUpdating
	_, err = c.swarmClient.K8slabV1alpha1().Swarms(namespace).UpdateStatus(ctx, updatedSwarm, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("unable to update swarm status map %v", err)
	}

	log.Infof("Update swarm %s namespace %s statefulset name %s configmap name %s version %d total workloads %d",
		sw.Name, sw.Spec.Namespace, sw.Spec.StatefulSetName, sw.Spec.ConfigMapName, sw.Spec.Version, len(sw.Spec.Workload))

	return nil
}

func (c *swarmController) statefulSetFromSwarm(namespace, name string) (*api.StatefulSet, error) {
	sw, err := c.swarmLister.Swarms(namespace).Get(name)
	if err != nil {
		return nil, fmt.Errorf("unable to get swarm namespace %s name %s", namespace, name)
	}

	log.Infof("Processing swarm %s namespace %s statefulset name %s configmap name %s version %d workloads %v",
		sw.Name, sw.Spec.Namespace, sw.Spec.StatefulSetName, sw.Spec.ConfigMapName, sw.Spec.Version, sw.Spec.Workload)

	sts, err := c.statefulSetLister.StatefulSets(sw.Spec.Namespace).Get(sw.Spec.StatefulSetName)
	if err != nil {
		return nil, fmt.Errorf("unable to get statefulset from swarm %s on namespace %s name %s error %v", sw.Name, sw.Spec.Namespace, sw.Spec.StatefulSetName, err)
	}
	return sts, nil
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

	ev := e.(event)
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

func (c *swarmController) handle(ctx context.Context, ev event) error {
	switch ev.action {
	case processSwarm:
		return c.process(ctx, ev.namespace, ev.name)
	case updatePool:
		return c.updatePool(ctx, ev.namespace, ev.name)
	}

	return fmt.Errorf("action %s not handled", ev.action)
}

func (c *swarmController) enqueue(a action, namespace, name string) {
	c.queue.Add(event{
		action:    a,
		namespace: namespace,
		name:      name,
	})
}
