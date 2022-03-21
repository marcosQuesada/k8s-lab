package app

import (
	"fmt"
	"time"
)

const DefaultWorkerFrequency = time.Second * 5

type State string

const (
	NeedsRefresh       State = "needs refresh"
	WaitingAssignation State = "waiting config"
	Syncing            State = "syncing"
)

type Event interface {
	Time() time.Time
	String() string
}

type stateEvent struct {
	time  time.Time
	state State
}

func newStateEvent(s State) *stateEvent {
	return &stateEvent{
		time:  time.Now(),
		state: s,
	}
}

func (e *stateEvent) Time() time.Time {
	return e.time
}

func (e *stateEvent) String() string {
	return fmt.Sprintf("%s got state %s", e.time, e.state)
}

type versionUpdateEvent struct {
	time    time.Time
	version int64
}

func newVersionUpdateEvent(v int64) *versionUpdateEvent {
	return &versionUpdateEvent{
		time:    time.Now(),
		version: v,
	}
}

func (e *versionUpdateEvent) Time() time.Time {
	return e.time
}

func (e *versionUpdateEvent) String() string {
	return fmt.Sprintf("%s got version %d", e.time, e.version)
}

type errorEvent struct {
	time  time.Time
	error error
}

func newErrorEvent(e error) *errorEvent {
	return &errorEvent{
		time:  time.Now(),
		error: e,
	}
}

func (e *errorEvent) Time() time.Time {
	return e.time
}

func (e *errorEvent) String() string {
	return fmt.Sprintf("%s got error %s", e.time, e.String())
}
