package client

import (
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

	// The only purpose of scraper name is to make runs easier for humans to identify
	// So, expect the run to start with the scraper name but there's probably more
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
