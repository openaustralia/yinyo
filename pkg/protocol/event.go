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
	e.RunID = jsonEvent.RunID
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
	case "first":
		var d FirstData
		err = json.Unmarshal(*jsonEvent.Data, &d)
		e.Data = d
	case "last":
		var d LastData
		err = json.Unmarshal(*jsonEvent.Data, &d)
		e.Data = d
	default:
		return errors.New("unexpected type")
	}
	return err
}

// NewLogEvent creates and returns a new log event
func NewLogEvent(id string, runID string, time time.Time, stage string, stream string, text string) Event {
	return Event{ID: id, RunID: runID, Time: time, Type: "log", Data: LogData{Stage: stage, Stream: stream, Text: text}}
}

// NewStartEvent creates and returns a new start event
func NewStartEvent(id string, runID string, time time.Time, stage string) Event {
	return Event{ID: id, RunID: runID, Time: time, Type: "start", Data: StartData{Stage: stage}}
}

// NewFinishEvent creates and returns a new finish event
func NewFinishEvent(id string, runID string, time time.Time, stage string, exitData ExitDataStage) Event {
	return Event{ID: id, RunID: runID, Time: time, Type: "finish", Data: FinishData{Stage: stage, ExitData: exitData}}
}

// NewFirstEvent creates and returns a new last event
func NewFirstEvent(id string, runID string, time time.Time) Event {
	return Event{ID: id, RunID: runID, Time: time, Type: "first", Data: FirstData{}}
}

// NewLastEvent creates and returns a new last event
func NewLastEvent(id string, runID string, time time.Time) Event {
	return Event{ID: id, RunID: runID, Time: time, Type: "last", Data: LastData{}}
}
