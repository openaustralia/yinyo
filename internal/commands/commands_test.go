package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStoragePath(t *testing.T) {
	assert.Equal(t, storagePath("abc", "app.tgz"), "abc/app.tgz")
	assert.Equal(t, storagePath("def", "output"), "def/output")
}

type MockJobDispatcher struct {
	mock.Mock
}

// TODO: Generate these mock methods automatically
func (m *MockJobDispatcher) CreateJobAndToken(namePrefix string, runToken string) (string, error) {
	args := m.Called(namePrefix, runToken)
	return args.String(0), args.Error(1)
}

func (m *MockJobDispatcher) StartJob(runName string, dockerImage string, command []string, env map[string]string) error {
	args := m.Called(runName, dockerImage, command, env)
	return args.Error(0)
}

func (m *MockJobDispatcher) DeleteJobAndToken(runName string) error {
	args := m.Called(runName)
	return args.Error(0)
}

func (m *MockJobDispatcher) GetToken(runName string) (string, error) {
	args := m.Called(runName)
	return args.String(0), args.Error(1)
}

func TestStartRun(t *testing.T) {
	// We want to mock the job dispatcher and check that it's called in the expected way
	job := new(MockJobDispatcher)
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
