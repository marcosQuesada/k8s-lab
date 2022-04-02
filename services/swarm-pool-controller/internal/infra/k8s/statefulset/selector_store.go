package statefulset

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	"sync"
)

type SelectorStore interface {
	EnsureRegister(namespace, name string, selector labels.Selector, swarmName string)
	UnRegister(namespace, name string)
	Matches(namespace, name string, l map[string]string) bool
	SwarmName(namespace, name string) (string, bool)
}

type item struct {
	selector  labels.Selector
	swarmName string
}

type selectorStore struct {
	index map[string]*item
	mutex sync.RWMutex
}

func NewSelectorStore() SelectorStore {
	return &selectorStore{
		index: map[string]*item{},
	}
}

func (s *selectorStore) EnsureRegister(namespace, name string, selector labels.Selector, swarmName string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	k := s.key(namespace, name)
	if _, ok := s.index[k]; ok {
		return
	}
	s.index[k] = &item{selector: selector, swarmName: swarmName}

	log.Infof("Registering key %s selector %s", k, selector.String())
}

func (s *selectorStore) UnRegister(namespace, name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	k := s.key(namespace, name)
	delete(s.index, k)
}

func (s *selectorStore) Matches(namespace, name string, l map[string]string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	k := s.key(namespace, name)
	sl, ok := s.index[k]
	if !ok {
		return false
	}

	return sl.selector.Matches(labels.Set(l))
}

func (s *selectorStore) SwarmName(namespace, name string) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	k := s.key(namespace, name)
	sl, ok := s.index[k]
	if !ok {
		return "", false
	}

	return sl.swarmName, true
}

func (s *selectorStore) key(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

func (s *selectorStore) len() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.index)
}
