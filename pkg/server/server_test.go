package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openaustralia/yinyo/internal/commands"
	"github.com/openaustralia/yinyo/pkg/blobstore"
	"github.com/openaustralia/yinyo/pkg/jobdispatcher"
	"github.com/stretchr/testify/assert"
)

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
	jobDispatcher := new(jobdispatcher.MockClient)
	app := &commands.App{BlobStore: blobstoreClient, JobDispatcher: jobDispatcher}
	server := Server{app: app}
	server.InitialiseRoutes()

	blobstoreClient.On("Get", "foo/app.tgz").Return(nil, errors.New("Doesn't exist"))
	blobstoreClient.On("IsNotExist", errors.New("Doesn't exist")).Return(true)
	jobDispatcher.On("GetToken", "foo").Return("abc123", nil)

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
	jobDispatcher.AssertExpectations(t)
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
