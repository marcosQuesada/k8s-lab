package app

type action string

const createSwarmAction = action("createSwarmAction")
const updateSwarmAction = action("updateSwarmAction")
const updateSwarmSizeAction = action("updateSwarmSizeAction")
const deleteSwarmAction = action("deleteSwarmAction")

type Event interface {
	Type() action
}

type createSwarm struct {
	namespace string
	name      string
}

func newCreateSwarm(namespace, name string) createSwarm {
	return createSwarm{namespace: namespace, name: name}
}

func (e createSwarm) Type() action {
	return createSwarmAction
}

type updateSwarm struct {
	namespace string
	name      string
}

func newUpdateSwarm(namespace, name string) updateSwarm {
	return updateSwarm{namespace: namespace, name: name}
}

func (e updateSwarm) Type() action {
	return updateSwarmAction
}

type updateSwarmSize struct {
	namespace string
	name      string
	size      int
}

func newUpdateSwarmSize(namespace, name string, size int) updateSwarmSize {
	return updateSwarmSize{namespace: namespace, name: name, size: size}
}

func (e updateSwarmSize) Type() action {
	return updateSwarmSizeAction
}

type deleteSwarm struct {
	namespace string
	name      string
}

func newDeleteSwarm(namespace, name string) deleteSwarm {
	return deleteSwarm{namespace: namespace, name: name}
}

func (e deleteSwarm) Type() action {
	return deleteSwarmAction
}
