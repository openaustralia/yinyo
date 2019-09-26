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

func defaultClient() Client {
	return NewClient("http://localhost:8080")
}

func TestCreateRun(t *testing.T) {
	client := defaultClient()
	run, err := client.CreateRun("foo")
	if err != nil {
		t.Fatal(err)
	}

	// The only purpose of name_prefix is to make runs easier for humans to identify
	// So, expect the run to start with the name_prefix but there's probably more
	assert.True(t, strings.HasPrefix(run.Name, "foo-"))
	assert.NotEqual(t, "", run.Token)
}

func TestCreateRunScraperNameEncoding(t *testing.T) {
	client := defaultClient()
	run, err := client.CreateRun("foo/b_12r")
	if err != nil {
		t.Fatal(err)
	}

	// Only certain characters are allowed in kubernetes job names
	assert.True(t, strings.HasPrefix(run.Name, "foo-b-12r-"))
}

// Check that run names are created to be unique even when the same scraper name
// is given twice
func TestCreateRunNamesUnique(t *testing.T) {
	client := defaultClient()
	run1, err := client.CreateRun("foo")
	if err != nil {
		t.Fatal(err)
	}
	run2, err := client.CreateRun("foo")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEqual(t, run1.Name, run2.Name)
}

func TestNamePrefixOptional(t *testing.T) {
	client := defaultClient()
	run, err := client.CreateRun("")
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, strings.HasPrefix(run.Name, "run-"))
}

func TestUploadDownloadApp(t *testing.T) {
	// First we need to create a run
	client := defaultClient()
	run, err := client.CreateRun("")
	if err != nil {
		t.Fatal(err)
	}
	// Now upload a random test pattern for the app
	app := "Random test pattern"
	body := strings.NewReader(app)
	resp, err := client.PutApp(run, body)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Now download the test pattern and check that it matches
	resp, err = client.GetApp(run)
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
