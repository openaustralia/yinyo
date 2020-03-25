package protocol

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// test marshalling and unmarshalling of usage
func testMarshalUsage(t *testing.T, usage Usage, jsonString string) {
	b, _ := json.Marshal(usage)
	assert.Equal(t, jsonString, string(b))
	var actual Usage
	_ = json.Unmarshal(b, &actual)
	assert.Equal(t, usage, actual)
}

func TestMarshalMemoryUsage(t *testing.T) {
	testMarshalUsage(t,
		// 50MB being used for 2 hours
		NewMemoryUsage("123", time.Date(2000, time.January, 2, 3, 45, 0, 0, time.UTC), 50*1024*1024, 2*60*60*1000*1000*1000),
		`{"time":"2000-01-02T03:45:00Z","run_id":"123","type":"memory","data":{"memory":52428800,"duration":7200}}`,
	)
}

func TestMarshalNetworkUsage(t *testing.T) {
	testMarshalUsage(t,
		NewNetworkUsage("123", time.Date(2000, time.January, 2, 3, 45, 0, 0, time.UTC), 50*1024*1024, 10*1024*1024),
		`{"time":"2000-01-02T03:45:00Z","run_id":"123","type":"network","data":{"in":52428800,"out":10485760}}`,
	)
}
