package operator

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type resourceEventHandler struct {
	queue workqueue.Interface
}

func NewResourceEventHandler(q workqueue.Interface) *resourceEventHandler {
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
	o := obj.(runtime.Object)
	r.queue.Add(&event{
		key: key,
		obj: o.DeepCopyObject(),
	})
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
	o := oldObj.(runtime.Object)
	n := newObj.(runtime.Object)
	r.queue.Add(&updateEvent{
		key:    key,
		newObj: n.DeepCopyObject(),
		oldObj: o.DeepCopyObject(),
	})
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
	o := obj.(runtime.Object)
	r.queue.Add(&event{
		key: key,
		obj: o.DeepCopyObject(),
	})
}
