package event

import (
	"encoding/json"
	"errors"
)

// JSONEvent is used for reading JSON
type JSONEvent struct {
	Type string           `json:"type"`
	Data *json.RawMessage `json:"data"`
}

// Event is the top level struct for representing events
type Event struct {
	Type string `json:"type"`
	Data Data   `json:"data"`
}

// Data is the interface for all core event data
type Data interface {
}

// StartData represents the start of a build or run
type StartData struct {
	Stage string `json:"stage"`
}

// FinishData represent the completion of a build or run
type FinishData struct {
	Stage string `json:"stage"`
}

// LogData is the output of some text from the build or run of a scraper
type LogData struct {
	Stage  string `json:"stage"`
	Stream string `json:"stream"`
	Text   string `json:"text"`
}

// LastData is the last event that's sent in a run
type LastData struct {
}

// UnmarshalJSON converts json to EventWrapper
func (e *Event) UnmarshalJSON(data []byte) error {
	var jsonEvent JSONEvent
	err := json.Unmarshal(data, &jsonEvent)
	if err != nil {
		return err
	}

	e.Type = jsonEvent.Type
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
func NewLogEvent(stage string, stream string, text string) Event {
	return Event{Type: "log", Data: LogData{Stage: stage, Stream: stream, Text: text}}
}

// NewStartEvent creates and returns a new start event
func NewStartEvent(stage string) Event {
	return Event{Type: "start", Data: StartData{Stage: stage}}
}

// NewFinishEvent creates and returns a new finish event
func NewFinishEvent(stage string) Event {
	return Event{Type: "finish", Data: FinishData{Stage: stage}}
}

// NewLastEvent creates and returns a new last event
func NewLastEvent() Event {
	return Event{Type: "last", Data: LastData{}}
}
