package protocol

import (
	"encoding/json"
	"errors"
	"time"
)

// UnmarshalJSON converts json to EventWrapper
func (e *Event) UnmarshalJSON(data []byte) error {
	var jsonEvent JSONEvent
	err := json.Unmarshal(data, &jsonEvent)
	if err != nil {
		return err
	}

	e.Type = jsonEvent.Type
	e.ID = jsonEvent.ID
	e.Time = jsonEvent.Time
	switch jsonEvent.Type {
	case "start":
		var d StartData
		err = json.Unmarshal(*jsonEvent.Data, &d)
		e.Data = d
	case "finish":
		var d FinishData
		err = json.Unmarshal(*jsonEvent.Data, &d)
		e.Data = d
	case "log":
		var d LogData
		err = json.Unmarshal(*jsonEvent.Data, &d)
		e.Data = d
	case "last":
		var d LastData
		err = json.Unmarshal(*jsonEvent.Data, &d)
		e.Data = d
	default:
		return errors.New("Unexpected type")
	}
	return err
}

// NewLogEvent creates and returns a new log event
func NewLogEvent(id string, time time.Time, stage string, stream string, text string) Event {
	return Event{ID: id, Time: time, Type: "log", Data: LogData{Stage: stage, Stream: stream, Text: text}}
}

// NewStartEvent creates and returns a new start event
func NewStartEvent(id string, time time.Time, stage string) Event {
	return Event{ID: id, Time: time, Type: "start", Data: StartData{Stage: stage}}
}

// NewFinishEvent creates and returns a new finish event
func NewFinishEvent(id string, time time.Time, stage string) Event {
	return Event{ID: id, Time: time, Type: "finish", Data: FinishData{Stage: stage}}
}

// NewLastEvent creates and returns a new last event
func NewLastEvent(id string, time time.Time) Event {
	return Event{ID: id, Time: time, Type: "last", Data: LastData{}}
}
