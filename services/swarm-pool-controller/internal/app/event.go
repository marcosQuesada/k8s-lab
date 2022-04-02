package app

type action string

const processSwarm = action("processSwarm")
const updatePool = action("updatePool")

type Event interface {
	Type() action
}

type processEvent struct {
	namespace string
	name      string
}

func newProcessEvent(namespace, name string) processEvent {
	return processEvent{namespace: namespace, name: name}
}

func (e processEvent) Type() action {
	return processSwarm
}

type updateEvent struct {
	namespace string
	name      string
	size      int
}

func newUpdateEvent(namespace, name string, size int) updateEvent {
	return updateEvent{namespace: namespace, name: name, size: size}
}

func (e updateEvent) Type() action {
	return updatePool
}
