package event

import (
	"encoding/json"
	"errors"
)

// JSONEvent is used for reading/writing JSON
type JSONEvent struct {
	Stage  string `json:"stage,omitempty"`
	Type   string `json:"type"`
	Stream string `json:"stream,omitempty"`
	Text   string `json:"text,omitempty"`
}

// Event is the top level struct for representing events
type Event struct {
	Data Data
}

// Data is the interface for all core event data
type Data interface {
}

// StartData represents the start of a build or run
type StartData struct {
	Stage string
}

// FinishData represent the completion of a build or run
type FinishData struct {
	Stage string
}

// LogData is the output of some text from the build or run of a scraper
type LogData struct {
	Stage  string
	Stream string
	Text   string
}

// LastData is the last event that's sent in a run
type LastData struct {
}

// MarshalJSON converts an EventWrapper to JSON
func (e Event) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Data)
}

// UnmarshalJSON converts json to EventWrapper
func (e *Event) UnmarshalJSON(data []byte) error {
	var jsonEvent JSONEvent
	err := json.Unmarshal(data, &jsonEvent)
	if err != nil {
		return err
	}

	switch jsonEvent.Type {
	case "start":
		e.Data = StartData{Stage: jsonEvent.Stage}
	case "finish":
		e.Data = FinishData{Stage: jsonEvent.Stage}
	case "log":
		e.Data = LogData{Stage: jsonEvent.Stage, Stream: jsonEvent.Stream, Text: jsonEvent.Text}
	case "last":
		e.Data = LastData{}
	default:
		return errors.New("Unexpected type")
	}
	return nil
}

// MarshalJSON converts a StartEvent to JSON
func (e StartData) MarshalJSON() ([]byte, error) {
	return json.Marshal(JSONEvent{Type: "start", Stage: e.Stage})
}

// MarshalJSON converts a FinishEvent to JSON
func (e FinishData) MarshalJSON() ([]byte, error) {
	return json.Marshal(JSONEvent{Type: "finish", Stage: e.Stage})
}

// MarshalJSON converts a LogEvent to JSON
func (e LogData) MarshalJSON() ([]byte, error) {
	return json.Marshal(JSONEvent{Type: "log", Stage: e.Stage, Stream: e.Stream, Text: e.Text})
}

// MarshalJSON converts a LastEvent to JSON
func (e LastData) MarshalJSON() ([]byte, error) {
	return json.Marshal(JSONEvent{Type: "last"})
}

// NewLogEvent creates and returns a new log event
func NewLogEvent(stage string, stream string, text string) Event {
	return Event{Data: LogData{Stage: stage, Stream: stream, Text: text}}
}

// NewStartEvent creates and returns a new start event
func NewStartEvent(stage string) Event {
	return Event{Data: StartData{Stage: stage}}
}

// NewFinishEvent creates and returns a new finish event
func NewFinishEvent(stage string) Event {
	return Event{Data: FinishData{Stage: stage}}
}

// NewLastEvent creates and returns a new last event
func NewLastEvent() Event {
	return Event{Data: LastData{}}
}
