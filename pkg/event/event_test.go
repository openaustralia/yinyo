package event

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// test marshalling and unmarshalling of event
func testMarshal(t *testing.T, event Event, jsonString string) {
	b, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, jsonString, string(b))
	var actual Event
	err = json.Unmarshal(b, &actual)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, event, actual)
}

func TestMarshalStartEvent(t *testing.T) {
	testMarshal(t,
		NewStartEvent("build"),
		`{"type":"start","data":{"stage":"build"}}`,
	)
}

func TestMarshalFinishEvent(t *testing.T) {
	testMarshal(t,
		NewFinishEvent("build"),
		`{"type":"finish","data":{"stage":"build"}}`,
	)
}

func TestMarshalLogEvent(t *testing.T) {
	testMarshal(t,
		NewLogEvent("build", "stdout", "Hello"),
		`{"type":"log","data":{"stage":"build","stream":"stdout","text":"Hello"}}`,
	)
}

func TestMarshalLastEvent(t *testing.T) {
	testMarshal(t,
		NewLastEvent(),
		`{"type":"last","data":{}}`,
	)
}
