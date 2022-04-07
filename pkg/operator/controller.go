package operator

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

type Handler interface {
	Create(ctx context.Context, o runtime.Object) error
	Update(ctx context.Context, o, n runtime.Object) error
	Delete(ctx context.Context, o runtime.Object) error
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

	eh := NewResourceEventHandler()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			o, err := eh.Create(obj)
			if err != nil {
				log.Errorf("unable to create, error %v", err)
				return
			}
			ctl.runner.Process(o)
		},
		UpdateFunc: func(old, new interface{}) {
			o, err := eh.Update(old, new)
			if err != nil {
				log.Errorf("unable to update, error %v", err)
				return
			}
			ctl.runner.Process(o)
		},
		DeleteFunc: func(obj interface{}) {
			o, err := eh.Delete(obj)
			if err != nil {
				log.Errorf("unable to delete, error %v", err)
				return
			}
			ctl.runner.Process(o)
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
	e, ok := k.(Event)
	if !ok {
		return fmt.Errorf("unexpected object type on handler, expected event got %T", k)
	}
	_, exists, err := c.informer.GetIndexer().GetByKey(e.GetKey())
	if err != nil {
		return fmt.Errorf("unable to fetching object with key %s from store: %v", e.GetKey(), err)
	}

	if !exists {
		log.Infof("handling deletion on key %s", e.GetKey())
		if ev, ok := e.(*event); ok {
			return c.eventHandler.Delete(ctx, ev.obj)
		}
		if ev, ok := e.(*updateEvent); ok {
			return c.eventHandler.Delete(ctx, ev.old)
		}
		return nil
	}

	switch ev := e.(type) {
	case *event:
		if e.GetAction() == Create {
			return c.eventHandler.Create(ctx, ev.obj)
		}
		return c.eventHandler.Delete(ctx, ev.obj)
	case *updateEvent:
		return c.eventHandler.Update(ctx, ev.old, ev.new)

	}

	return fmt.Errorf("unexpected object type on handler, expected Event got %T", e)
}
