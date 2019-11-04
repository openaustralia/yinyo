package test

// This tests the clay-run executable without running it in a kubernetes cluster

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"github.com/openaustralia/morph-ng/pkg/clayclient"
	"github.com/stretchr/testify/assert"
)

func checkRequest(t *testing.T, r *http.Request, method string, path string, body string) {
	assert.Equal(t, method, r.Method)
	assert.Equal(t, path, r.URL.EscapedPath())
	b, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, body, string(b))
}

func TestSimpleRun(t *testing.T) {
	count := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		if count == 0 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				"{\"stage\":\"build\",\"type\":\"started\"}",
			)
		} else if count == 1 {
			checkRequest(t, r, "GET", "/runs/run-name/app", "")
			w.Header().Set("Content-Type", "application/gzip")
			reader, err := clayclient.CreateArchiveFromDirectory("fixtures/scrapers/hello-world")
			if err != nil {
				t.Fatal(err)
			}
			_, err = io.Copy(w, reader)
			if err != nil {
				t.Fatal(err)
			}
		} else if count == 2 {
			checkRequest(t, r, "GET", "/runs/run-name/cache", "")
			// Let the client know that there is no cache in this case
			http.NotFound(w, r)
		} else if count == 3 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				"{\"stage\":\"build\",\"type\":\"log\",\"stream\":\"stdout\",\"text\":\"Built\"}",
			)
		} else if count == 4 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				"{\"stage\":\"build\",\"type\":\"finished\"}",
			)
		} else if count == 5 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				"{\"stage\":\"run\",\"type\":\"started\"}",
			)
		} else if count == 6 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				"{\"stage\":\"run\",\"type\":\"log\",\"stream\":\"stdout\",\"text\":\"Ran\"}",
			)
		} else if count == 7 {
			checkRequest(t, r,
				"PUT",
				"/runs/run-name/exit-data",
				"{\"exit_code\": 0, \"usage\": {\"build\": {}, \"run\": {}}}\n",
				// "{\"exit_code\":0,\"usage\":{\"build\":{\"wall_time\":0,\"cpu_time\":0,\"max_rss\":0,\"network_in\":0,\"network_out\":0},\"run\":{\"wall_time\":0,\"cpu_time\":0,\"max_rss\":0,\"network_in\":0,\"network_out\":0}}}",
			)
		} else if count == 8 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				"{\"stage\":\"run\",\"type\":\"finished\"}",
			)
		} else if count == 9 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				"EOF",
			)
		} else {
			fmt.Println("Didn't expect so many requests")
			t.Fatal("Didn't expect so many requests")
		}
		count++
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	// Just run it and see what breaks
	cmd := exec.Command("/bin/bash", "../../build/package/run.sh", "run-name", "output.txt")
	cmd.Env = append(os.Environ(),
		// Send requests for the clay server to our local test server instead (which we start here)
		"CLAY_INTERNAL_SERVER_URL="+ts.URL,
		"CLAY_INTERNAL_BUILD_COMMAND=echo Built",
		"CLAY_INTERNAL_RUN_COMMAND=echo Ran",
	)

	stdoutStderr, err := cmd.CombinedOutput()
	fmt.Printf("%s\n", stdoutStderr)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Test that app is correctly downloaded
	// TODO: Test that cache is correctly downloaded
	// TODO: Test that cache is correctly uploaded
	// TODO: Test that output is correctly uploaded
}
