package operator

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

// ErrNoAppLabelFound it happens on label not found on resource
var ErrNoAppLabelFound = errors.New("no app label found ")

type labelSelectorMiddleware struct {
	label        string
	eventHandler EventHandler
}

func NewLabelSelectorMiddleware(watchedLabel string, eh EventHandler) EventHandler {
	return &labelSelectorMiddleware{
		label:        watchedLabel,
		eventHandler: eh,
	}
}

func (f *labelSelectorMiddleware) Add(obj interface{}) {
	o, ok := obj.(runtime.Object)
	if !ok {
		return
	}
	if !f.isMatched(o) {
		return
	}
	f.eventHandler.Add(o)
}

func (f *labelSelectorMiddleware) Update(oldObj, newObj interface{}) {
	o, ok := oldObj.(runtime.Object)
	if !ok {
		return
	}
	n, ok := newObj.(runtime.Object)
	if !ok {
		return
	}
	if !f.isMatched(o) {
		return
	}
	f.eventHandler.Update(o, n)
}

func (f *labelSelectorMiddleware) Delete(obj interface{}) {
	o, ok := obj.(runtime.Object)
	if !ok {
		return
	}
	if !f.isMatched(o) {
		return
	}
	f.eventHandler.Delete(o)
}

func (f *labelSelectorMiddleware) isMatched(obj runtime.Object) bool {
	v, err := f.hasWatchedLabel(obj)
	if err != nil {
		log.Debugf("unable to get object %T app label, error %v", obj, err)
		return false
	}

	return v
}

func (f *labelSelectorMiddleware) hasWatchedLabel(obj runtime.Object) (bool, error) {
	label, err := f.getAppLabel(obj)
	if err != nil {
		return false, fmt.Errorf("meta app label error: %v", err)
	}

	return label == f.label, nil
}

func (f *labelSelectorMiddleware) getAppLabel(obj runtime.Object) (string, error) {
	acc, err := meta.Accessor(obj)
	if err != nil {
		return "", fmt.Errorf("meta accessor error: %v", err)
	}

	labels := acc.GetLabels()
	v, ok := labels["app"]
	if !ok {
		return "", ErrNoAppLabelFound
	}

	return v, nil
}
