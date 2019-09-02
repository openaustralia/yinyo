package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// First go at doing end-to-end testing of the system
func TestServerHello(t *testing.T) {
	// This assumes the API is accessible at http://localhost:8080
	resp, err := http.Get("http://localhost:8080")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, []string{"text/plain; charset=utf-8"}, resp.Header["Content-Type"])
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "Hello from Clay!\n", string(b))
}

// TODO: Put the tests in a different package so no chance that we could use things we probably shouldn't
type createRunResult struct {
	RunName  string `json:"run_name"`
	RunToken string `json:"run_token"`
}

func TestCreateRun(t *testing.T) {
	resp, err := http.Post("http://localhost:8080/runs?scraper_name=foo", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	decoder := json.NewDecoder(resp.Body)
	var result createRunResult
	err = decoder.Decode(&result)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEqual(t, "", result.RunName)
	assert.NotEqual(t, "", result.RunToken)
}
