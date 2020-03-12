package protocol

import (
	"encoding/json"
	"errors"
	"math"
	"time"
)

// Usage is the top level struct for representing usage that's reported externally
type Usage struct {
	Time  time.Time   `json:"time"`
	RunID string      `json:"run_id"`
	Type  string      `json:"type"`
	Data  interface{} `json:"data"`
}

// JSONUsage is used for reading JSON
type JSONUsage struct {
	Time  time.Time        `json:"time"`
	RunID string           `json:"run_id"`
	Type  string           `json:"type"`
	Data  *json.RawMessage `json:"data"`
}

// NewMemoryUsage creates and returns a new memory usage
func NewMemoryUsage(runID string, time time.Time, memory uint64, duration time.Duration) Usage {
	// Round up to the nearest second
	return Usage{RunID: runID, Time: time, Type: "memory", Data: MemoryUsageData{Memory: memory, Duration: uint64(math.Ceil(duration.Seconds()))}}
}

// NewNetworkUsage creates and returns a new network usage
func NewNetworkUsage(runID string, time time.Time, in uint64, out uint64) Usage {
	return Usage{RunID: runID, Time: time, Type: "network", Data: NetworkUsageData{In: in, Out: out}}
}

type MemoryUsageData struct {
	Memory   uint64 `json:"memory"`
	Duration uint64 `json:"duration"`
}

type NetworkUsageData struct {
	In  uint64 `json:"in"`
	Out uint64 `json:"out"`
}

// UnmarshalJSON converts json to EventWrapper
func (usage *Usage) UnmarshalJSON(data []byte) error {
	var jsonUsage JSONUsage
	err := json.Unmarshal(data, &jsonUsage)
	if err != nil {
		return err
	}

	usage.Type = jsonUsage.Type
	usage.RunID = jsonUsage.RunID
	usage.Time = jsonUsage.Time

	switch jsonUsage.Type {
	case "memory":
		var d MemoryUsageData
		err = json.Unmarshal(*jsonUsage.Data, &d)
		usage.Data = d
	case "network":
		var d NetworkUsageData
		err = json.Unmarshal(*jsonUsage.Data, &d)
		usage.Data = d
	default:
		return errors.New("unexpected type")
	}
	return err
}
