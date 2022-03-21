package operator

import "k8s.io/apimachinery/pkg/runtime"

// Event wraps k8s updated keys
type Event interface {
	GetKey() string
}

type event struct {
	key string
	obj runtime.Object
}

// GetKey returns event key
func (e *event) GetKey() string {
	return e.key
}

type updateEvent struct {
	key    string
	oldObj runtime.Object
	newObj runtime.Object
}

// GetKey returns event key
func (e *updateEvent) GetKey() string {
	return e.key
}
