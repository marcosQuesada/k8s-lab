package operator

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

var ErrNoAppLabelFound = errors.New("no app label found ")

type selector struct {
	label string
}

func NewSelector(watchedLabel string) Selector {
	return &selector{
		label: watchedLabel,
	}
}

func (f *selector) Validate(obj runtime.Object) error {
	v, err := f.hasWatchedLabel(obj)
	if err != nil {
		log.Debugf("unable to get object %T app label, error %v", obj, err)
		return err
	}

	if !v {
		l, err := f.getAppLabel(obj)
		if err != nil {
			return fmt.Errorf("selector validate error, not watched label, app label not found, skip object, error %v", err)
		}
		return fmt.Errorf("label %s not queue", l)
	}

	return nil
}

func (f *selector) hasWatchedLabel(obj runtime.Object) (bool, error) {
	label, err := f.getAppLabel(obj)
	if err != nil {
		return false, fmt.Errorf("meta app label error: %v", err)
	}

	return label == f.label, nil
}

func (f *selector) getAppLabel(obj runtime.Object) (string, error) {
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

type nopValidator struct{}

func NewNopValidator() Selector {
	return &nopValidator{}
}

func (n *nopValidator) Validate(obj runtime.Object) error {
	return nil
}
