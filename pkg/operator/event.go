package operator

import "k8s.io/apimachinery/pkg/runtime"

type Action string

const Create = Action("create")
const Update = Action("update")
const Delete = Action("delete")

// Event wraps k8s updated keys
type Event interface {
	GetKey() string
	GetAction() Action
}

type event struct {
	key    string
	obj    runtime.Object
	action Action
}

func newCreateEvent(key string, obj runtime.Object) Event {
	return &event{
		key:    key,
		obj:    obj.DeepCopyObject(),
		action: Create,
	}
}

func newDeleteEvent(key string, obj runtime.Object) Event {
	return &event{
		key:    key,
		obj:    obj.DeepCopyObject(),
		action: Delete,
	}
}

// GetKey returns event key
func (e *event) GetKey() string {
	return e.key
}

// GetAction returns action type
func (e *event) GetAction() Action {
	return e.action
}

type updateEvent struct {
	key string
	old runtime.Object
	new runtime.Object
}

func newUpdateEvent(key string, old, new runtime.Object) Event {
	return &updateEvent{
		key: key,
		old: old.DeepCopyObject(),
		new: new.DeepCopyObject(),
	}
}

// GetKey returns event key
func (e *updateEvent) GetKey() string {
	return e.key
}

// GetAction returns action type
func (e *updateEvent) GetAction() Action {
	return Update
}
