package operator

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

var errNilObject = errors.New("nil object received")

// EventHandler gets called on each event variation from informer
type EventHandler interface {
	Create(obj interface{}) (Event, error)
	Update(old, new interface{}) (Event, error)
	Delete(obj interface{}) (Event, error)
}

type resourceEventHandler struct {
}

func NewResourceEventHandler() EventHandler {
	return &resourceEventHandler{}
}

// Create add object to the queue on valid label
func (r *resourceEventHandler) Create(obj interface{}) (Event, error) {
	if obj == nil {
		return nil, errNilObject
	}

	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		return nil, fmt.Errorf("create MetaNamespaceKeyFunc error %v", err)
	}

	log.Debugf("Add %T: %s", obj, key)
	return newCreateEvent(key, obj.(runtime.Object)), nil
}

// Update object to the queue on valid label
func (r *resourceEventHandler) Update(o, n interface{}) (Event, error) {
	if o == nil || n == nil {
		log.Errorf("Update with Nil Object old: %T new: %T", o, n)
		return nil, errNilObject
	}

	key, err := cache.MetaNamespaceKeyFunc(n)
	if err != nil {
		return nil, fmt.Errorf("update MetaNamespaceKeyFunc error %v", err)
	}

	// @TODO: Should be recover event struct ?
	//diff := cmp.Diff(old, new)
	//cleanDiff := strings.TrimFunc(diff, func(r rune) bool {
	//	return !unicode.IsGraphic(r)
	//})
	//fmt.Printf("UPDATE %s diff: %s \n", resourceType, cleanDiff)

	log.Debugf("Update %T: %s", o, key)
	return newUpdateEvent(key, o.(runtime.Object), n.(runtime.Object)), nil
}

// Delete object to the queue on valid label
func (r *resourceEventHandler) Delete(obj interface{}) (Event, error) {
	if obj == nil {
		log.Error("Delete with nil obj, skip")
		return nil, errNilObject
	}

	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		return nil, fmt.Errorf("delete DeletionHandlingMetaNamespaceKeyFunc error %v", err)
	}

	log.Debugf("Delete %T: %s", obj, key)
	return newDeleteEvent(key, obj.(runtime.Object)), nil
}
