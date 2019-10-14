package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/openaustralia/morph-ng/pkg/jobdispatcher"
	"github.com/openaustralia/morph-ng/pkg/stream"
)

func TestStoragePath(t *testing.T) {
	assert.Equal(t, storagePath("abc", "app.tgz"), "abc/app.tgz")
	assert.Equal(t, storagePath("def", "output"), "def/output")
}

func TestStartRun(t *testing.T) {
	// We want to mock the job dispatcher and check that it's called in the expected way
	job := new(jobdispatcher.MockClient)
	job.On(
		"StartJob",
		"run-name",
		"openaustralia/clay-scraper:v1",
		[]string{"/bin/run.sh", "run-name", "output.txt"},
		map[string]string{"FOO": "bar", "CLAY_INTERNAL_RUN_TOKEN": "supersecret"},
	).Return(nil)
	job.On("GetToken", "run-name").Return("supersecret", nil)

	app := App{Job: job}
	err := app.StartRun("run-name", "output.txt", map[string]string{"FOO": "bar"})
	assert.Nil(t, err)
	job.AssertExpectations(t)
}

// Make sure that setting a reserved environment variable is not allowed
func TestStartRunWithReservedEnv(t *testing.T) {
	app := App{}
	err := app.StartRun("run-name", "output.txt", map[string]string{"CLAY_INTERNAL_FOO": "bar"})
	assert.EqualError(t, err, "Can't override environment variables starting with CLAY_INTERNAL_")
}

func TestCreateEvent(t *testing.T) {
	stream := new(stream.MockStream)
	stream.On("Add", "run-name", "{\"some\": \"json\"}").Return(nil)

	app := App{Stream: stream}
	err := app.CreateEvent("run-name", "{\"some\": \"json\"}")
	assert.Nil(t, err)
}
