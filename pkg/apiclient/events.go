package apiclient

// Utilities to make it just a little easier to create different types of events

import (
	"time"

	"github.com/openaustralia/yinyo/pkg/protocol"
)

// CreateStartEvent creates and sends a "start" event
func (run *Run) CreateStartEvent(stage string) (int, error) {
	return run.CreateEvent(protocol.NewStartEvent("", time.Now(), stage))
}

// CreateFinishEvent creates and sends a "finish" event
func (run *Run) CreateFinishEvent(stage string, exitData protocol.ExitDataStage) (int, error) {
	return run.CreateEvent(protocol.NewFinishEvent("", time.Now(), stage, exitData))
}

func (run *Run) CreateNetworkEvent(in uint64, out uint64) (int, error) {
	return run.CreateEvent(protocol.NewNetworkEvent("", time.Now(), in, out))
}

// CreateLogEvent creates and sends a "log" event
func (run *Run) CreateLogEvent(stage string, stream string, text string) (int, error) {
	return run.CreateEvent(protocol.NewLogEvent("", time.Now(), stage, stream, text))
}

// CreateLastEvent creates and sends a "last" event
func (run *Run) CreateLastEvent() (int, error) {
	return run.CreateEvent(protocol.NewLastEvent("", time.Now()))
}
