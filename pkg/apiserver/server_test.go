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
// Use "" for the token if you don't want the request to be authenticated
func makeRequest(app commands.App, method string, url string, body io.Reader, token string) *httptest.ResponseRecorder {
	server := Server{app: app, maxRunTime: 86400}
	server.InitialiseRoutes()

	req, _ := http.NewRequest(method, url, body)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)
	return rr
}

func TestCreateRun(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("CreateRun", "").Return(protocol.Run{Name: "run-foo", Token: "dsfhg"}, nil)

	rr := makeRequest(app, "POST", "/runs", nil, "")

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t,
		`{"name":"run-foo","token":"dsfhg"}
`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json"}}, rr.Header())
	app.AssertExpectations(t)
}

func TestCreateRunWithPrefix(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("CreateRun", "prefix").Return(protocol.Run{Name: "prefix-foo", Token: "abc"}, nil)

	rr := makeRequest(app, "POST", "/runs?name_prefix=prefix", nil, "")

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t,
		`{"name":"prefix-foo","token":"abc"}
`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json"}}, rr.Header())
	app.AssertExpectations(t)
}

func TestCreateRunInternalServerError(t *testing.T) {
	app := new(commandsmocks.App)
	// There was some kind of internal error when creating a run
	app.On("CreateRun", "").Return(protocol.Run{}, errors.New("Something internal"))

	rr := makeRequest(app, "POST", "/runs", nil, "")

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, `{"error":"Internal server error"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.Header())
	app.AssertExpectations(t)
}

// Make sure that when we call start with a bad json body we get a sensible error
func TestStartBadBody(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("GetTokenCache", "foo").Return("abc123", nil)

	rr := makeRequest(app, "POST", "/runs/foo/start", strings.NewReader(`{"env":"foo"}`), "abc123")

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"error":"JSON in body not correctly formatted"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.Header())
	app.AssertExpectations(t)
}

// If we haven't uploaded an app error when starting a run
func TestStartNoApp(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("GetTokenCache", "foo").Return("abc123", nil)
	app.On("StartRun", "foo", "", map[string]string{}, "", int64(86400)).Return(commands.ErrAppNotAvailable)

	rr := makeRequest(app, "POST", "/runs/foo/start", strings.NewReader(`{}`), "abc123")

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"error":"app needs to be uploaded before starting a run"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.Header())
	app.AssertExpectations(t)
}

func TestStart(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("GetTokenCache", "foo").Return("abc123", nil)
	app.On("StartRun", "foo", "", map[string]string{}, "", int64(86400)).Return(nil)

	rr := makeRequest(app, "POST", "/runs/foo/start", strings.NewReader("{}"), "abc123")

	assert.Equal(t, http.StatusOK, rr.Code)

	app.AssertExpectations(t)
}

func TestStartLowerMaxRunTime(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("GetTokenCache", "foo").Return("abc123", nil)
	app.On("StartRun", "foo", "", map[string]string{}, "", int64(120)).Return(nil)

	rr := makeRequest(app, "POST", "/runs/foo/start", strings.NewReader(`{"max_run_time": 120}`), "abc123")

	assert.Equal(t, http.StatusOK, rr.Code)

	app.AssertExpectations(t)
}

func TestStartHigherMaxRunTime(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("GetTokenCache", "foo").Return("abc123", nil)

	rr := makeRequest(app, "POST", "/runs/foo/start", strings.NewReader(`{"max_run_time": 100000}`), "abc123")

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"error":"max_run_time should not be larger than 86400"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.Header())

	app.AssertExpectations(t)
}

func TestCreateEventBadBody(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("GetTokenCache", "foo").Return("abc123", nil)

	rr := makeRequest(app, "POST", "/runs/foo/events", strings.NewReader(`{"event":"broken"}`), "abc123")

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"error":"JSON in body not correctly formatted"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.Header())
	app.AssertExpectations(t)
}

func TestPutApp(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("GetTokenCache", "run-name").Return("abc123", nil)
	app.On("PutApp", mock.Anything, int64(3), "run-name").Return(nil)

	rr := makeRequest(app, "PUT", "/runs/run-name/app", strings.NewReader("foo"), "abc123")

	assert.Equal(t, http.StatusOK, rr.Code)
	app.AssertExpectations(t)
}

func TestPutAppWrongRunName(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("GetTokenCache", "does-not-exist").Return("", commands.ErrNotFound)

	rr := makeRequest(app, "PUT", "/runs/does-not-exist/app", strings.NewReader(""), "abc123")

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, `{"error":"run does-not-exist: not found"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.Header())
	app.AssertExpectations(t)
}

func TestGetApp(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("GetTokenCache", "my-run").Return("abc123", nil)
	app.On("GetApp", "my-run").Return(strings.NewReader("code stuff"), nil)

	rr := makeRequest(app, "GET", "/runs/my-run/app", nil, "abc123")

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "code stuff", rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/gzip"}}, rr.Header())
	app.AssertExpectations(t)
}

// This tests if the app isn't found (rather than the run)
func TestGetAppErrNotFound(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("GetTokenCache", "my-run").Return("abc123", nil)
	app.On("GetApp", "my-run").Return(nil, commands.ErrNotFound)

	rr := makeRequest(app, "GET", "/runs/my-run/app", nil, "abc123")

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, `{"error":"not found"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.Header())
	app.AssertExpectations(t)
}

func TestGetAppNoBearerToken(t *testing.T) {
	app := new(commandsmocks.App)

	rr := makeRequest(app, "GET", "/runs/my-run/app", nil, "")

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Equal(t, `{"error":"expected Authorization header with bearer token"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.Header())
	app.AssertExpectations(t)
}

func TestGetAppBadToken(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("GetTokenCache", "my-run").Return("def456", nil)

	rr := makeRequest(app, "GET", "/runs/my-run/app", nil, "abc123")

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Equal(t, `{"error":"authorization header has incorrect bearer token"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.Header())
	app.AssertExpectations(t)
}

func TestGetCache(t *testing.T) {
	app := new(commandsmocks.App)

	app.On("GetTokenCache", "my-run").Return("abc123", nil)
	app.On("GetCache", "my-run").Return(strings.NewReader("cached stuff"), nil)

	rr := makeRequest(app, "GET", "/runs/my-run/cache", nil, "abc123")

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "cached stuff", rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/gzip"}}, rr.Header())
	app.AssertExpectations(t)
}

func TestPutCache(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("GetTokenCache", "my-run").Return("abc123", nil)
	app.On("PutCache", mock.Anything, int64(12), "my-run").Return(nil)

	rr := makeRequest(app, "PUT", "/runs/my-run/cache", strings.NewReader("cached stuff"), "abc123")

	assert.Equal(t, http.StatusOK, rr.Code)
	app.AssertExpectations(t)
}

func TestGetOutput(t *testing.T) {
	app := new(commandsmocks.App)

	app.On("GetTokenCache", "my-run").Return("abc123", nil)
	app.On("GetOutput", "my-run").Return(strings.NewReader("output stuff"), nil)

	rr := makeRequest(app, "GET", "/runs/my-run/output", nil, "abc123")

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "output stuff", rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/octet-stream"}}, rr.Header())
	app.AssertExpectations(t)
}

func TestPutOutput(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("GetTokenCache", "my-run").Return("abc123", nil)
	app.On("PutOutput", mock.Anything, int64(12), "my-run").Return(nil)

	rr := makeRequest(app, "PUT", "/runs/my-run/output", strings.NewReader("output stuff"), "abc123")

	assert.Equal(t, http.StatusOK, rr.Code)
	app.AssertExpectations(t)
}

func TestGetExitData(t *testing.T) {
	app := new(commandsmocks.App)
	exitData := protocol.ExitData{
		Build: &protocol.ExitDataStage{
			ExitCode: 12,
		},
	}
	app.On("GetTokenCache", "my-run").Return("abc123", nil)
	app.On("GetExitData", "my-run").Return(exitData, nil)

	rr := makeRequest(app, "GET", "/runs/my-run/exit-data", nil, "abc123")

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "{\"build\":{\"exit_code\":12,\"usage\":{\"wall_time\":0,\"cpu_time\":0,\"max_rss\":0,\"network_in\":0,\"network_out\":0}}}\n", rr.Body.String())
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
		protocol.NewStartEvent("", time, "build"),
		protocol.NewFinishEvent("", time, "build", protocol.ExitDataStage{}),
	}}
	app.On("GetTokenCache", "my-run").Return("abc123", nil)
	app.On("GetEvents", "my-run", "0").Return(events)
	rr := makeRequest(app, "GET", "/runs/my-run/events", nil, "abc123")

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{"time":"2000-01-02T03:45:00Z","type":"start","data":{"stage":"build"}}
{"time":"2000-01-02T03:45:00Z","type":"finish","data":{"stage":"build","exit_data":{"exit_code":0,"usage":{"wall_time":0,"cpu_time":0,"max_rss":0,"network_in":0,"network_out":0}}}}
`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/ld+json"}}, rr.Header())

	app.AssertExpectations(t)
}

func TestDelete(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("GetTokenCache", "my-run").Return("abc123", nil)
	app.On("DeleteRun", "my-run").Return(nil)

	rr := makeRequest(app, "DELETE", "/runs/my-run", nil, "abc123")

	assert.Equal(t, http.StatusOK, rr.Code)
	app.AssertExpectations(t)
}

func TestHello(t *testing.T) {
	app := new(commandsmocks.App)

	rr := makeRequest(app, "GET", "/", nil, "")

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "Hello from Yinyo!\n", rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}}, rr.Header())
	app.AssertExpectations(t)
}
