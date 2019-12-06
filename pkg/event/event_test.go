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
		`{"stage":"build","type":"start"}`,
	)
}

func TestMarshalFinishEvent(t *testing.T) {
	testMarshal(t,
		NewFinishEvent("build"),
		`{"stage":"build","type":"finish"}`,
	)
}

func TestMarshalLogEvent(t *testing.T) {
	testMarshal(t,
		NewLogEvent("build", "stdout", "Hello"),
		`{"stage":"build","type":"log","stream":"stdout","text":"Hello"}`,
	)
}

func TestMarshalLastEvent(t *testing.T) {
	testMarshal(t,
		NewLastEvent(),
		`{"type":"last"}`,
	)
}
