package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
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

func TestCreateRun(t *testing.T) {
	result, err := createRun("foo")
	if err != nil {
		t.Fatal(err)
	}

	// The only purpose of name_prefix is to make runs easier for humans to identify
	// So, expect the run to start with the name_prefix but there's probably more
	assert.True(t, strings.HasPrefix(result.RunName, "foo-"))
	assert.NotEqual(t, "", result.RunToken)
}

func TestCreateRunScraperNameEncoding(t *testing.T) {
	result, err := createRun("foo/b_12r")
	if err != nil {
		t.Fatal(err)
	}

	// Only certain characters are allowed in kubernetes job names
	assert.True(t, strings.HasPrefix(result.RunName, "foo-b-12r-"))
}

// Check that run names are created to be unique even when the same scraper name
// is given twice
func TestCreateRunNamesUnique(t *testing.T) {
	result1, err := createRun("foo")
	if err != nil {
		t.Fatal(err)
	}
	result2, err := createRun("foo")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEqual(t, result1.RunName, result2.RunName)
}

func TestNamePrefixOptional(t *testing.T) {
	result, err := createRun("")
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, strings.HasPrefix(result.RunName, "run-"))
}

func TestUploadDownloadApp(t *testing.T) {
	// First we need to create a run
	run, err := createRun("")
	if err != nil {
		t.Fatal(err)
	}
	// Now upload a random test pattern for the app
	app := "Random test pattern"
	body := strings.NewReader(app)
	url := fmt.Sprintf("http://localhost:8080/runs/%s/app", run.RunName)
	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+run.RunToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Now download the test pattern and check that it matches
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+run.RunToken)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, []string{"application/gzip"}, resp.Header["Content-Type"])
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, app, string(b))
	// TODO: Clean up run
}
