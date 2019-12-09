package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Make sure that when we call start with a bad json body we get a sensible error
func TestStartBadBody(t *testing.T) {
	req, err := http.NewRequest("POST", "/runs/foo/start", strings.NewReader(`{"env":"foo"}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := appHandler(start)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"error":"JSON in body not correctly formatted"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.HeaderMap)
}

func TestCreateEventBadBody(t *testing.T) {
	req, err := http.NewRequest("POST", "/runs/foo/events", strings.NewReader(`{"event":"broken"}`))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := appHandler(createEvent)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, `{"error":"JSON in body not correctly formatted"}`, rr.Body.String())
	assert.Equal(t, http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}, rr.HeaderMap)
}
