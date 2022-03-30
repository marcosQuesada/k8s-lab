package controller

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"time"
)

const maxRetries = 5

type Handler interface {
	Create(ctx context.Context, o runtime.Object) error
	Remove(ctx context.Context, namespace, name string) error
}

type Controller struct {
	queue        workqueue.RateLimitingInterface
	informer     cache.SharedIndexInformer
	eventHandler Handler
	resourceType string
}

func NewController(eventHandler Handler, informer cache.SharedIndexInformer, resourceType string) *Controller {
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
			ctl.enqueue(new)
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err != nil {
				log.Errorf("Unable to Delete event from object %T error %v", obj, err)
				return
			}
			log.Infof("Processing delete to %v: %s", resourceType, key)
			queue.Add(key)
		},
	})

	return ctl
}

func (c *Controller) Run(ctx context.Context) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()
	go c.informer.Run(ctx.Done())

	if !cache.WaitForNamedCacheSync(c.resourceType, ctx.Done(), c.informer.HasSynced) {
		return
	}

	go wait.UntilWithContext(ctx, c.worker, time.Second)
}

func (c *Controller) worker(ctx context.Context) {
	for c.processNextItem(ctx) {
	}
}

func (c *Controller) processNextItem(ctx context.Context) bool {
	k, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(k)

	key := k.(string)
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
		return c.eventHandler.Remove(ctx, namespace, name)

	}
	o, ok := obj.(runtime.Object)
	if !ok {
		return fmt.Errorf("unexpected object type on handler, expected runtime object got %T", obj)
	}

	return c.eventHandler.Create(ctx, o)
}

func (c *Controller) enqueue(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		log.Errorf("Unable to Update event from object %T error %v", obj, err)
		utilruntime.HandleError(err)
		return
	}
	log.Infof("enqueue to %T: %s", obj, key)
	c.queue.Add(key)
}
