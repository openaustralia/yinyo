package protocol

import (
	"encoding/json"
	"time"
)

// All the types here are used in the yinyo API. So, they all will get serialised and deserialised.
// Therefore, for all types include an explicit instruction for JSON marshalling/unmarshalling.

type CreateRunOptions struct {
	APIKey string
}

// StartRunOptions are options that can be used when starting a run
type StartRunOptions struct {
	Output     string        `json:"output"`
	Callback   Callback      `json:"callback"`
	Env        []EnvVariable `json:"env"`
	MaxRunTime int64         `json:"max_run_time"`
	Memory     int64         `json:"memory"`
}

// Callback represents what we need to know to make a particular callback request
// This is not just a string so that we could support adding headers or other special things
// in the callback request
type Callback struct {
	URL string `json:"url"`
}

// EnvVariable is the name and value of an environment variable
type EnvVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ExitData holds information about how things ran and how much resources were used
type ExitData struct {
	Build    *ExitDataStage `json:"build,omitempty"`
	Execute  *ExitDataStage `json:"execute,omitempty"`
	Finished bool           `json:"finished"`
}

// ExitDataStage gives the exit data for a single stage
type ExitDataStage struct {
	ExitCode int        `json:"exit_code"`
	Usage    StageUsage `json:"usage"`
}

// Usage gives the resource usage for a single stage
type StageUsage struct {
	MaxRSS     uint64 `json:"max_rss"`     // In bytes
	NetworkIn  uint64 `json:"network_in"`  // In bytes
	NetworkOut uint64 `json:"network_out"` // In bytes
}

// Run is what you get when you create a run and what you need to update it
type Run struct {
	ID string `json:"id"`
}

// JSONEvent is used for reading JSON
type JSONEvent struct {
	ID    string           `json:"id"`
	RunID string           `json:"run_id"`
	Time  time.Time        `json:"time"`
	Type  string           `json:"type"`
	Data  *json.RawMessage `json:"data"`
}

// Event is the top level struct for representing events
type Event struct {
	ID    string    `json:"id,omitempty"`
	RunID string    `json:"run_id"`
	Time  time.Time `json:"time"`
	Type  string    `json:"type"`
	Data  Data      `json:"data"`
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
	Stage    string        `json:"stage"`
	ExitData ExitDataStage `json:"exit_data"`
}

// LogData is the output of some text from the build or run of a scraper
type LogData struct {
	Stage  string `json:"stage"`
	Stream string `json:"stream"`
	Text   string `json:"text"`
}

// FirstData is the first event that's sent in a run
type FirstData struct {
}

// LastData is the last event that's sent in a run
type LastData struct {
}

// Hello gives some basic useful information about the server
type Hello struct {
	Message     string        `json:"message"`
	MaxRunTime  DefaultAndMax `json:"max_run_time"`
	Memory      DefaultAndMax `json:"memory"`
	Version     string        `json:"version"`
	RunnerImage string        `json:"runner_image"`
}

// DefaultAndMax is a useful type for int64 resources that have a default and a maximum value
type DefaultAndMax struct {
	Default int64 `json:"default"`
	Max     int64 `json:"max"`
}
