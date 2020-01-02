package apiserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openaustralia/yinyo/internal/commands"
	"github.com/openaustralia/yinyo/pkg/blobstore"
	"github.com/openaustralia/yinyo/pkg/keyvaluestore"
	"github.com/stretchr/testify/assert"
)

func TestCreateRunInternalServerError(t *testing.T) {
	req, err := http.NewRequest("POST", "/runs", nil)
	if err != nil {
		t.Fatal(err)
	}
	app := new(commands.MockApp)
	// There was some kind of internal error when creating a run
	app.On("CreateRun", "").Return(commands.CreateRunResult{}, errors.New("Something internal"))
	server := Server{app: app}
	server.InitialiseRoutes()

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, `{"error":"Internal server error"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.HeaderMap)
	app.AssertExpectations(t)
}

// Make sure that when we call start with a bad json body we get a sensible error
func TestStartBadBody(t *testing.T) {
	req, err := http.NewRequest("POST", "/runs/foo/start", strings.NewReader(`{"env":"foo"}`))
	if err != nil {
		t.Fatal(err)
	}

	server := Server{}

	rr := httptest.NewRecorder()
	handler := appHandler(server.start)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"error":"JSON in body not correctly formatted"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.HeaderMap)
}

// If we haven't uploaded an app error when starting a run
func TestStartNoApp(t *testing.T) {
	blobstoreClient := new(blobstore.MockClient)
	keyvalueStore := new(keyvaluestore.MockClient)
	app := &commands.AppImplementation{BlobStore: blobstoreClient, KeyValueStore: keyvalueStore}
	server := Server{app: app}
	server.InitialiseRoutes()

	blobstoreClient.On("Get", "foo/app.tgz").Return(nil, errors.New("Doesn't exist"))
	blobstoreClient.On("IsNotExist", errors.New("Doesn't exist")).Return(true)
	keyvalueStore.On("Get", "token:foo").Return("abc123", nil)

	req, err := http.NewRequest("POST", "/runs/foo/start", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer abc123")

	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"error":"app needs to be uploaded before starting a run"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.HeaderMap)
	blobstoreClient.AssertExpectations(t)
	keyvalueStore.AssertExpectations(t)
}

func TestCreateEventBadBody(t *testing.T) {
	req, err := http.NewRequest("POST", "/runs/foo/events", strings.NewReader(`{"event":"broken"}`))
	if err != nil {
		t.Fatal(err)
	}

	server := Server{}
	rr := httptest.NewRecorder()
	handler := appHandler(server.createEvent)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"error":"JSON in body not correctly formatted"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.HeaderMap)
}

func TestPutAppWrongRunName(t *testing.T) {
	app := new(commands.MockApp)
	app.On("GetTokenCache", "does-not-exist").Return("", commands.ErrNotFound)

	server := Server{app: app}
	server.InitialiseRoutes()

	req, err := http.NewRequest("PUT", "/runs/does-not-exist/app", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer abc123")

	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, `{"error":"run does-not-exist: not found"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.HeaderMap)

	app.AssertExpectations(t)
}

// This tests if the app isn't found (rather than the run)
func TestGetAppErrNotFound(t *testing.T) {
	app := new(commands.MockApp)
	app.On("GetTokenCache", "my-run").Return("abc123", nil)
	app.On("GetApp", "my-run").Return(nil, commands.ErrNotFound)
	server := Server{app: app}
	server.InitialiseRoutes()

	req, err := http.NewRequest("GET", "/runs/my-run/app", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer abc123")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Equal(t, `{"error":"not found"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.HeaderMap)
	app.AssertExpectations(t)
}

func TestGetAppNoBearerToken(t *testing.T) {
	server := Server{}
	server.InitialiseRoutes()

	req, err := http.NewRequest("GET", "/runs/my-run/app", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Equal(t, `{"error":"Expected Authorization header with bearer token"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.HeaderMap)
}

func TestGetAppBadToken(t *testing.T) {
	app := new(commands.MockApp)
	app.On("GetTokenCache", "my-run").Return("def456", nil)
	server := Server{app: app}
	server.InitialiseRoutes()

	req, err := http.NewRequest("GET", "/runs/my-run/app", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer abc123")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Equal(t, `{"error":"Authorization header has incorrect bearer token"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.HeaderMap)
	app.AssertExpectations(t)
}
