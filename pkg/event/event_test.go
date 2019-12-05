package event

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshalStartEvent(t *testing.T) {
	b, err := json.Marshal(NewStartEvent("build"))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, `{"stage":"build","type":"start"}`, string(b))
}

func TestMarshalFinishEvent(t *testing.T) {
	b, err := json.Marshal(NewFinishEvent("build"))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, `{"stage":"build","type":"finish"}`, string(b))
}

func TestMarshalLogEvent(t *testing.T) {
	b, err := json.Marshal(NewLogEvent("build", "stdout", "Hello"})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, `{"stage":"build","type":"log","stream":"stdout","text":"Hello"}`, string(b))
}
