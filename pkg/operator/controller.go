package operator

import (
	"context"
	"fmt"
	"github.com/google/go-cmp/cmp"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"strings"
	"time"
	"unicode"
)

const maxRetries = 5
const handleTimeout = time.Second
const workerFrequency = time.Second

type Handler interface {
	Handle(ctx context.Context, o runtime.Object) error
	HandleDeletion(ctx context.Context, namespace, name string) error
}

type Controller struct {
	queue        workqueue.RateLimitingInterface
	informer     cache.SharedIndexInformer
	eventHandler Handler
	resourceType string
}

func New(eventHandler Handler, informer cache.SharedIndexInformer, resourceType string) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	ctl := &Controller{
		informer:     informer,
		queue:        queue,
		eventHandler: eventHandler,
		resourceType: resourceType,
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ctl.enqueue(obj)
		},
		UpdateFunc: func(old, new interface{}) {
			// @TODO: Should be recover event struct ?
			diff := cmp.Diff(old, new)
			cleanDiff := strings.TrimFunc(diff, func(r rune) bool {
				return !unicode.IsGraphic(r)
			})
			fmt.Printf("UPDATE %s diff: %s \n", resourceType, cleanDiff)

			ctl.enqueue(new)
		},
		DeleteFunc: func(obj interface{}) {
			ctl.enqueue(obj)
		},
	})

	return ctl
}

func (c *Controller) Run(ctx context.Context) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	if !cache.WaitForNamedCacheSync(c.resourceType, ctx.Done(), c.informer.HasSynced) {
		return
	}

	log.Infof("%s First Cache Synced on version %s", c.resourceType, c.informer.LastSyncResourceVersion())

	wait.UntilWithContext(ctx, c.worker, workerFrequency)
}

func (c *Controller) worker(ctx context.Context) {
	for c.processNextItem(ctx) {
	}
}

func (c *Controller) processNextItem(ctx context.Context) bool {
	k, quit := c.queue.Get()
	if quit {
		log.Infof("Queue from %s goes down!", c.resourceType)
		return false
	}
	defer c.queue.Done(k)

	log.Infof("%s Cache Synced on version %s", c.resourceType, c.informer.LastSyncResourceVersion())

	key := k.(string)
	ctx, cancel := context.WithTimeout(ctx, handleTimeout)
	defer cancel()

	err := c.handle(ctx, key)
	if err == nil || errors.HasStatusCause(err, v1.NamespaceTerminatingCause) {
		c.queue.Forget(k)
		return true
	}

	if c.queue.NumRequeues(k) < maxRetries {
		log.Errorf("Error processing key %s, retry. Error: %v", key, err)
		c.queue.AddRateLimited(k)
		return true
	}

	log.Errorf("Error processing %s Max retries achieved: %v", key, err)
	c.queue.Forget(k)
	utilruntime.HandleError(err)

	return true
}

func (c *Controller) handle(ctx context.Context, key string) error {
	obj, exists, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return fmt.Errorf("unable to fetching object with key %s from store: %v", key, err)
	}

	if !exists {
		namespace, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			return fmt.Errorf("unable to split  object with key %s from store: %v", key, err)
		}
		log.Infof("handling deletion on key %s", key)
		return c.eventHandler.HandleDeletion(ctx, namespace, name)
	}

	o, ok := obj.(runtime.Object)
	if !ok {
		return fmt.Errorf("unexpected object type on handler, expected runtime object got %T", obj)
	}

	return c.eventHandler.Handle(ctx, o)
}

func (c *Controller) enqueue(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		log.Errorf("Unable to Update event from object %T error %v", obj, err)
		utilruntime.HandleError(err)
		return
	}
	log.Debugf("enqueue %T: %s", obj, key)
	c.queue.Add(key)
}