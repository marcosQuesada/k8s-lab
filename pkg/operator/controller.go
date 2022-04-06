package operator

import (
	"context"
	"fmt"
	"github.com/google/go-cmp/cmp"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"strings"
	"unicode"
)

type Handler interface {
	Handle(ctx context.Context, o runtime.Object) error // @TODO: Split into Create/Update to remove handler caches
	Delete(ctx context.Context, namespace, name string) error
}

type Controller struct {
	runner       Runner
	informer     cache.SharedIndexInformer
	eventHandler Handler
	resourceType string
}

func New(eventHandler Handler, informer cache.SharedIndexInformer, runner Runner, resourceType string) *Controller {
	ctl := &Controller{
		informer:     informer,
		runner:       runner,
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

	if !cache.WaitForNamedCacheSync(c.resourceType, ctx.Done(), c.informer.HasSynced) {
		return
	}

	log.Infof("%s First Cache Synced on version %s", c.resourceType, c.informer.LastSyncResourceVersion())

	c.runner.Run(ctx, c.handle)
}

func (c *Controller) handle(ctx context.Context, k interface{}) error {
	key := k.(string)
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
		return c.eventHandler.Delete(ctx, namespace, name)
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
	c.runner.Process(key)
}
