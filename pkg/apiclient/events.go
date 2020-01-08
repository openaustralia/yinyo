package apiclient

// Utilities to make it just a little easier to create different types of events

import (
	"time"

	"github.com/openaustralia/yinyo/pkg/protocol"
)

func (run *Run) CreateStartEvent(stage string) error {
	return run.CreateEvent(protocol.NewStartEvent("", time.Now(), stage))
}

func (run *Run) CreateFinishEvent(stage string) error {
	return run.CreateEvent(protocol.NewFinishEvent("", time.Now(), stage))
}

func (run *Run) CreateLogEvent(stage string, stream string, text string) error {
	return run.CreateEvent(protocol.NewLogEvent("", time.Now(), stage, stream, text))
}

func (run *Run) CreateLastEvent() error {
	return run.CreateEvent(protocol.NewLastEvent("", time.Now()))
}
