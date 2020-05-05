package apiserver

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	commandsmocks "github.com/openaustralia/yinyo/mocks/pkg/commands"
	"github.com/openaustralia/yinyo/pkg/commands"
	"github.com/openaustralia/yinyo/pkg/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Makes a request to the server and records the response for testing purposes
func makeRequest(app commands.App, method string, url string, body io.Reader) *httptest.ResponseRecorder {
	server := Server{app: app, defaultMaxRunTime: 3600, maxRunTime: 86400, defaultMemory: 1073741824, maxMemory: 1610612736, version: "development", runDockerImage: "openaustralia/yinyo-runner:abc"}
	server.InitialiseRoutes()

	req, _ := http.NewRequest(method, url, body)
	// Make the request come "internally"
	req.RemoteAddr = "10.0.0.1:11111"

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)
	return rr
}

func TestCreateRun(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("CreateRun", protocol.CreateRunOptions{}).Return(protocol.Run{ID: "run-foo"}, nil)

	rr := makeRequest(app, "POST", "/runs", nil)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t,
		`{"id":"run-foo"}
`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json"}}, rr.Header())
	app.AssertExpectations(t)
}

func TestCreateRunInternalServerError(t *testing.T) {
	app := new(commandsmocks.App)
	// There was some kind of internal error when creating a run
	app.On("CreateRun", protocol.CreateRunOptions{}).Return(protocol.Run{}, errors.New("Something internal"))

	rr := makeRequest(app, "POST", "/runs", nil)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, `{"error":"Internal server error"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.Header())
	app.AssertExpectations(t)
}

// Make sure that when we call start with a bad json body we get a sensible error
func TestStartBadBody(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("IsRunCreated", "foo").Return(true, nil)

	rr := makeRequest(app, "POST", "/runs/foo/start", strings.NewReader(`{"env":"foo"}`))

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"error":"JSON in body not correctly formatted"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.Header())
	app.AssertExpectations(t)
}

// If we haven't uploaded an app error when starting a run
func TestStartNoApp(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("IsRunCreated", "foo").Return(true, nil)
	app.On("StartRun", "foo", "openaustralia/yinyo-runner:abc", protocol.StartRunOptions{MaxRunTime: 3600, Memory: 1073741824}).Return(commands.ErrAppNotAvailable)

	rr := makeRequest(app, "POST", "/runs/foo/start", strings.NewReader(`{}`))

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"error":"app needs to be uploaded before starting a run"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.Header())
	app.AssertExpectations(t)
}

func TestStartWithDefaults(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("IsRunCreated", "foo").Return(true, nil)
	app.On("StartRun", "foo", "openaustralia/yinyo-runner:abc", protocol.StartRunOptions{MaxRunTime: 3600, Memory: 1073741824}).Return(nil)

	rr := makeRequest(app, "POST", "/runs/foo/start", strings.NewReader("{}"))

	assert.Equal(t, http.StatusOK, rr.Code)

	app.AssertExpectations(t)
}

func TestStartLowerMaxRunTime(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("IsRunCreated", "foo").Return(true, nil)
	app.On("StartRun", "foo", "openaustralia/yinyo-runner:abc", protocol.StartRunOptions{MaxRunTime: 120, Memory: 1073741824}).Return(nil)

	rr := makeRequest(app, "POST", "/runs/foo/start", strings.NewReader(`{"max_run_time": 120}`))

	assert.Equal(t, http.StatusOK, rr.Code)

	app.AssertExpectations(t)
}

func TestStartHigherMaxRunTime(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("IsRunCreated", "foo").Return(true, nil)

	rr := makeRequest(app, "POST", "/runs/foo/start", strings.NewReader(`{"max_run_time": 100000}`))

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"error":"max_run_time should not be larger than 86400"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.Header())

	app.AssertExpectations(t)
}

func TestStartLowerMaxMemory(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("IsRunCreated", "foo").Return(true, nil)
	app.On("StartRun", "foo", "openaustralia/yinyo-runner:abc", protocol.StartRunOptions{MaxRunTime: 3600, Memory: 1024}).Return(nil)

	rr := makeRequest(app, "POST", "/runs/foo/start", strings.NewReader(`{"memory": 1024}`))

	assert.Equal(t, http.StatusOK, rr.Code)

	app.AssertExpectations(t)
}

func TestStartHigherMaxMemory(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("IsRunCreated", "foo").Return(true, nil)

	rr := makeRequest(app, "POST", "/runs/foo/start", strings.NewReader(`{"memory": 2147483648}`))

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"error":"memory should not be larger than 1610612736"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.Header())

	app.AssertExpectations(t)
}

func TestCreateEventBadBody(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("IsRunCreated", "foo").Return(true, nil)

	rr := makeRequest(app, "POST", "/runs/foo/events", strings.NewReader(`{"event":"broken"}`))

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"error":"JSON in body not correctly formatted"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.Header())
	app.AssertExpectations(t)
}

func TestPutApp(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("IsRunCreated", "run-name").Return(true, nil)
	app.On("PutApp", "run-name", mock.Anything, int64(3)).Return(nil)

	rr := makeRequest(app, "PUT", "/runs/run-name/app", strings.NewReader("foo"))

	assert.Equal(t, http.StatusOK, rr.Code)
	app.AssertExpectations(t)
}

func TestPutAppWrongRunName(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("IsRunCreated", "does-not-exist").Return(false, nil)

	rr := makeRequest(app, "PUT", "/runs/does-not-exist/app", strings.NewReader(""))

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, `{"error":"run does-not-exist: not found"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.Header())
	app.AssertExpectations(t)
}

func TestGetApp(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("IsRunCreated", "my-run").Return(true, nil)
	app.On("GetApp", "my-run").Return(strings.NewReader("code stuff"), nil)

	rr := makeRequest(app, "GET", "/runs/my-run/app", nil)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "code stuff", rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/gzip"}}, rr.Header())
	app.AssertExpectations(t)
}

// This tests if the app isn't found (rather than the run)
func TestGetAppErrNotFound(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("IsRunCreated", "my-run").Return(true, nil)
	app.On("GetApp", "my-run").Return(nil, commands.ErrNotFound)

	rr := makeRequest(app, "GET", "/runs/my-run/app", nil)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, `{"error":"not found"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.Header())
	app.AssertExpectations(t)
}

func TestGetCache(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("IsRunCreated", "my-run").Return(true, nil)
	app.On("GetCache", "my-run").Return(strings.NewReader("cached stuff"), nil)

	rr := makeRequest(app, "GET", "/runs/my-run/cache", nil)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "cached stuff", rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/gzip"}}, rr.Header())
	app.AssertExpectations(t)
}

func TestPutCache(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("IsRunCreated", "my-run").Return(true, nil)
	app.On("PutCache", "my-run", mock.Anything, int64(12)).Return(nil)

	rr := makeRequest(app, "PUT", "/runs/my-run/cache", strings.NewReader("cached stuff"))

	assert.Equal(t, http.StatusOK, rr.Code)
	app.AssertExpectations(t)
}

func TestGetOutput(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("IsRunCreated", "my-run").Return(true, nil)
	app.On("GetOutput", "my-run").Return(strings.NewReader("output stuff"), nil)

	rr := makeRequest(app, "GET", "/runs/my-run/output", nil)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "output stuff", rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/octet-stream"}}, rr.Header())
	app.AssertExpectations(t)
}

func TestPutOutput(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("IsRunCreated", "my-run").Return(true, nil)
	app.On("PutOutput", "my-run", mock.Anything, int64(12)).Return(nil)

	rr := makeRequest(app, "PUT", "/runs/my-run/output", strings.NewReader("output stuff"))

	assert.Equal(t, http.StatusOK, rr.Code)
	app.AssertExpectations(t)
}

func TestGetExitData(t *testing.T) {
	app := new(commandsmocks.App)
	exitData := protocol.ExitData{
		Build: &protocol.ExitDataStage{
			ExitCode: 12,
		},
		Finished: true,
	}
	app.On("IsRunCreated", "my-run").Return(true, nil)
	app.On("GetExitData", "my-run").Return(exitData, nil)

	rr := makeRequest(app, "GET", "/runs/my-run/exit-data", nil)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"build":{"exit_code":12,"usage":{"max_rss":0,"network_in":0,"network_out":0}},"finished":true}
`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json"}}, rr.Header())
	app.AssertExpectations(t)
}

// Make a fake event iterator that we can use for testing
type events struct {
	contents []protocol.Event
	index    int
}

func (e *events) More() bool {
	return e.index < len(e.contents)
}

func (e *events) Next() (protocol.Event, error) {
	event := e.contents[e.index]
	e.index++
	return event, nil
}

func TestGetEvents(t *testing.T) {
	app := new(commandsmocks.App)
	time := time.Date(2000, time.January, 2, 3, 45, 0, 0, time.UTC)
	events := &events{contents: []protocol.Event{
		protocol.NewStartEvent("", "abc", time, "build"),
		protocol.NewFinishEvent("", "abc", time, "build", protocol.ExitDataStage{}),
	}}
	app.On("IsRunCreated", "my-run").Return(true, nil)
	app.On("GetEvents", "my-run", "0").Return(events)

	rr := makeRequest(app, "GET", "/runs/my-run/events", nil)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"run_id":"abc","time":"2000-01-02T03:45:00Z","type":"start","data":{"stage":"build"}}
{"run_id":"abc","time":"2000-01-02T03:45:00Z","type":"finish","data":{"stage":"build","exit_data":{"exit_code":0,"usage":{"max_rss":0,"network_in":0,"network_out":0}}}}
`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/ld+json"}}, rr.Header())

	app.AssertExpectations(t)
}

func TestDelete(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("IsRunCreated", "my-run").Return(true, nil)
	app.On("DeleteRun", "my-run").Return(nil)

	rr := makeRequest(app, "DELETE", "/runs/my-run", nil)

	assert.Equal(t, http.StatusOK, rr.Code)
	app.AssertExpectations(t)
}

func TestHello(t *testing.T) {
	app := new(commandsmocks.App)

	rr := makeRequest(app, "GET", "/", nil)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"message":"Hello from Yinyo!","max_run_time":{"default":3600,"max":86400},"memory":{"default":1073741824,"max":1610612736},"version":"development","runner_image":"openaustralia/yinyo-runner:abc"}
`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json"}}, rr.Header())

	app.AssertExpectations(t)
}
