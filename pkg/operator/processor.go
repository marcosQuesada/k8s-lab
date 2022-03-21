package operator

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

// ListWatcher defines list and watch methods
type ListWatcher interface {
	List(options metav1.ListOptions) (runtime.Object, error)
	Watch(options metav1.ListOptions) (watch.Interface, error)
}

// ResourceHandler gets called on each resource variation consumed from queue
type ResourceHandler interface {
	Created(ctx context.Context, obj runtime.Object)
	Updated(ctx context.Context, old, new runtime.Object)
	Deleted(ctx context.Context, obj runtime.Object)
}

// EventHandler gets called on each event variation from informer
type EventHandler interface {
	Add(obj interface{})
	Update(oldObj, newObj interface{})
	Delete(obj interface{})
}

type eventProcessor struct {
	indexer  cache.Indexer
	informer cache.Controller
	handler  ResourceHandler
}

// NewEventProcessor instantiates EventProcessor
func NewEventProcessor(o runtime.Object, w ListWatcher, eh EventHandler, h ResourceHandler) *eventProcessor {
	indexer, informer := cache.NewIndexerInformer(
		&cache.ListWatch{
			ListFunc:  w.List,
			WatchFunc: w.Watch,
		},
		o,
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    eh.Add,
			UpdateFunc: eh.Update,
			DeleteFunc: eh.Delete,
		},
		cache.Indexers{},
	)

	return &eventProcessor{
		indexer:  indexer,
		informer: informer,
		handler:  h,
	}
}

// Run eventProcessor, sync informer before proceed
func (e *eventProcessor) Run(stopCh chan struct{}) {
	go e.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, e.informer.HasSynced) {
		utilruntime.HandleError(errors.New("timed out waiting for caches to sync"))
		return
	}
}

// Handle update event
func (e *eventProcessor) Handle(ctx context.Context, ev Event) error {
	_, exists, err := e.indexer.GetByKey(ev.GetKey())
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", ev.GetKey(), err)
		return err
	}

	if !exists {
		log.Infof("Object %s does not exist anymore, delete! event: %T", ev.GetKey(), e)
		if ev, ok := ev.(*event); ok {
			e.handler.Deleted(ctx, ev.obj)
		}
		if ev, ok := ev.(*updateEvent); ok {
			e.handler.Deleted(ctx, ev.oldObj)
		}
		return nil
	}

	switch ev := ev.(type) {
	case *event:
		e.handler.Created(ctx, ev.obj)
	case *updateEvent:
		e.handler.Updated(ctx, ev.newObj, ev.oldObj)
	}

	return nil
}
