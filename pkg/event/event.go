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

// EventWrapper is the top level struct for representing events
type EventWrapper struct {
	Event Event
}

// Event is the interface for all event types
type Event interface {
}

// StartEvent represents the start of a build or run
type StartEvent struct {
	Stage string
}

// FinishEvent represent the completion of a build or run
type FinishEvent struct {
	Stage string
}

// LogEvent is the output of some text from the build or run of a scraper
type LogEvent struct {
	Stage  string
	Stream string
	Text   string
}

// LastEvent is the last event that's sent in a run
type LastEvent struct {
}

// MarshalJSON converts an EventWrapper to JSON
func (e EventWrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.Event)
}

// UnmarshalJSON converts json to EventWrapper
func (e *EventWrapper) UnmarshalJSON(data []byte) error {
	var jsonEvent JSONEvent
	err := json.Unmarshal(data, &jsonEvent)
	if err != nil {
		return err
	}

	switch jsonEvent.Type {
	case "start":
		e.Event = StartEvent{Stage: jsonEvent.Stage}
	case "finish":
		e.Event = FinishEvent{Stage: jsonEvent.Stage}
	case "log":
		e.Event = LogEvent{Stage: jsonEvent.Stage, Stream: jsonEvent.Stream, Text: jsonEvent.Text}
	case "last":
		e.Event = LastEvent{}
	default:
		return errors.New("Unexpected type")
	}
	return nil
}

// MarshalJSON converts a StartEvent to JSON
func (e StartEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(JSONEvent{Type: "start", Stage: e.Stage})
}

// MarshalJSON converts a FinishEvent to JSON
func (e FinishEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(JSONEvent{Type: "finish", Stage: e.Stage})
}

// MarshalJSON converts a LogEvent to JSON
func (e LogEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(JSONEvent{Type: "log", Stage: e.Stage, Stream: e.Stream, Text: e.Text})
}

// MarshalJSON converts a LastEvent to JSON
func (e LastEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(JSONEvent{Type: "last"})
}
