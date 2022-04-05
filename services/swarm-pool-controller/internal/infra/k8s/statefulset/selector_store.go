package statefulset

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"sync"
)

type SelectorStore interface {
	Register(namespace, name string, ls *metav1.LabelSelector) error
	UnRegister(namespace, name string)
	Matches(namespace, name string, l map[string]string) bool
	IsRegistered(namespace, name string) bool
}

type selectorStore struct {
	index map[string]labels.Selector
	mutex sync.RWMutex
}

func NewSelectorStore() SelectorStore {
	return &selectorStore{
		index: map[string]labels.Selector{},
	}
}

func (s *selectorStore) Register(namespace, name string, ls *metav1.LabelSelector) error {
	selector, err := metav1.LabelSelectorAsSelector(ls)
	if err != nil {
		return fmt.Errorf("unable to get label selector, error %v", err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	k := namespace + "/" + name
	if _, ok := s.index[k]; ok {
		return nil
	}
	s.index[k] = selector

	log.Infof("Registering key %s selector %s", k, selector.String())
	return nil
}

func (s *selectorStore) UnRegister(namespace, name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.index, namespace+"/"+name)
}

func (s *selectorStore) Matches(namespace, name string, l map[string]string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	sl, ok := s.index[namespace+"/"+name]
	if !ok {
		return false
	}

	return sl.Matches(labels.Set(l))
}

func (s *selectorStore) IsRegistered(namespace, name string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	_, ok := s.index[namespace+"/"+name]
	return ok
}

func (s *selectorStore) len() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.index)
}
