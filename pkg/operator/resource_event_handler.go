package operator

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// EventHandler gets called on each event variation from informer
type EventHandler interface {
	Add(obj interface{})
	Update(oldObj, newObj interface{})
	Delete(obj interface{})
}

type resourceEventHandler struct {
	queue workqueue.Interface
}

func NewResourceEventHandler(q workqueue.Interface) EventHandler {
	return &resourceEventHandler{
		queue: q,
	}
}

// Add object to the queue on valid label
func (r *resourceEventHandler) Add(obj interface{}) {
	if obj == nil {
		log.Error("Add with nil obj, skip")
		return
	}

	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		log.Errorf("Add MetaNamespaceKeyFunc error %v", err)
		return
	}

	log.Debugf("Add %T: %s", obj, key)
	r.queue.Add(key)
}

// Update object to the queue on valid label
func (r *resourceEventHandler) Update(oldObj, newObj interface{}) {
	if newObj == nil || oldObj == nil {
		log.Errorf("Update with Nil Object old: %T new: %T", oldObj, newObj)
		return
	}

	key, err := cache.MetaNamespaceKeyFunc(newObj)
	if err != nil {
		log.Errorf("Patch MetaNamespaceKeyFunc error %v", err)
		return
	}

	log.Debugf("Update %T: %s", oldObj, key)
	r.queue.Add(key)
}

// Delete object to the queue on valid label
func (r *resourceEventHandler) Delete(obj interface{}) {
	if obj == nil {
		log.Error("Delete with nil obj, skip")
		return
	}

	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		log.Errorf("Delete DeletionHandlingMetaNamespaceKeyFunc error %v", err)
		return
	}

	log.Debugf("Delete %T: %s", obj, key)
	r.queue.Add(key)
}
