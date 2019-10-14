package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/openaustralia/morph-ng/pkg/jobdispatcher"
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
		map[string]string{"FOO": "bar"},
	).Return(nil)
	app := App{Job: job}
	app.StartRun("run-name", "output.txt", map[string]string{"FOO": "bar"})
	job.AssertExpectations(t)
}
