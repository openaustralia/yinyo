package apiserver

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	commandsmocks "github.com/openaustralia/yinyo/mocks/pkg/commands"
	"github.com/openaustralia/yinyo/pkg/commands"
	"github.com/openaustralia/yinyo/pkg/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Makes a request to the server and records the response for testing purposes
// Use "" for the token if you don't want the request to be authenticated
func makeRequest(app commands.App, method string, url string, body io.Reader, token string) *httptest.ResponseRecorder {
	server := Server{app: app}
	server.InitialiseRoutes()

	req, _ := http.NewRequest(method, url, body)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)
	return rr
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
	app.On("StartRun", "foo", "", map[string]string{}, "").Return(commands.ErrAppNotAvailable)

	rr := makeRequest(app, "POST", "/runs/foo/start", strings.NewReader(`{}`), "abc123")

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"error":"app needs to be uploaded before starting a run"}`, rr.Body.String())
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
	assert.Equal(t, `{"error":"Expected Authorization header with bearer token"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.Header())
	app.AssertExpectations(t)
}

func TestGetAppBadToken(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("GetTokenCache", "my-run").Return("def456", nil)

	rr := makeRequest(app, "GET", "/runs/my-run/app", nil, "abc123")

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Equal(t, `{"error":"Authorization header has incorrect bearer token"}`, rr.Body.String())
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

func TestPutExitData(t *testing.T) {
	app := new(commandsmocks.App)
	app.On("GetTokenCache", "my-run").Return("abc123", nil)
	exitData := protocol.ExitData{
		Build: &protocol.ExitDataStage{
			ExitCode: 12,
		},
	}
	app.On("PutExitData", "my-run", exitData).Return(nil)

	rr := makeRequest(app, "PUT", "/runs/my-run/exit-data", strings.NewReader(`{"build":{"exit_code":12,"usage":{"wall_time":0,"cpu_time":0,"max_rss":0,"network_in":0,"network_out":0}}}`), "abc123")

	assert.Equal(t, http.StatusOK, rr.Code)
	app.AssertExpectations(t)
}
