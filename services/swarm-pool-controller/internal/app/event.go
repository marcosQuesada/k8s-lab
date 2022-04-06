package app

type action string

const processSwarmAction = action("processSwarmAction")
const updateSwarmAction = action("updateSwarmAction")
const deleteSwarmAction = action("deleteSwarmAction")

type Event interface {
	Type() action
}

type processSwarm struct {
	namespace string
	name      string
}

func newProcessSwarm(namespace, name string) processSwarm {
	return processSwarm{namespace: namespace, name: name}
}

func (e processSwarm) Type() action {
	return processSwarmAction
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
	return updateSwarmAction
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
