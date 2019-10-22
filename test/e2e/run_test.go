package test

// This tests the run.sh script without running it in a kubernetes cluster

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleRun(t *testing.T) {
	count := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		if count == 0 {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/runs/run-name/events", r.URL.EscapedPath())
			body, err := ioutil.ReadAll(r.Body)
			r.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "{\"stage\":\"build\",\"type\":\"started\"}", string(body))
			// fmt.Fprintln(w, "Hello, client")
		} else if count == 1 {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/runs/run-name/app", r.URL.EscapedPath())
		} else if count == 2 {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/runs/run-name/cache", r.URL.EscapedPath())
		} else if count == 3 {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/runs/run-name/events", r.URL.EscapedPath())
			body, err := ioutil.ReadAll(r.Body)
			r.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "{\"stage\":\"build\",\"type\":\"log\",\"stream\":\"stdout\",\"text\":\"Built\"}", string(body))
		} else if count == 4 {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/runs/run-name/events", r.URL.EscapedPath())
			body, err := ioutil.ReadAll(r.Body)
			r.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "{\"stage\":\"build\",\"type\":\"finished\"}", string(body))
		} else if count == 5 {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/runs/run-name/events", r.URL.EscapedPath())
			body, err := ioutil.ReadAll(r.Body)
			r.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "{\"stage\":\"run\",\"type\":\"started\"}", string(body))
		} else if count == 6 {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/runs/run-name/events", r.URL.EscapedPath())
			body, err := ioutil.ReadAll(r.Body)
			r.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "{\"stage\":\"run\",\"type\":\"log\",\"stream\":\"stdout\",\"text\":\"Ran\"}", string(body))
		} else if count == 7 {
			assert.Equal(t, "PUT", r.Method)
			assert.Equal(t, "/runs/run-name/exit-data", r.URL.EscapedPath())
			body, err := ioutil.ReadAll(r.Body)
			r.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "{\"exit_code\": 0, \"usage\": {\"build\": {}, \"run\": {}}}\n", string(body))
		} else if count == 8 {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/runs/run-name/events", r.URL.EscapedPath())
			body, err := ioutil.ReadAll(r.Body)
			r.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "{\"stage\":\"run\",\"type\":\"finished\"}", string(body))
		} else if count == 9 {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/runs/run-name/events", r.URL.EscapedPath())
			body, err := ioutil.ReadAll(r.Body)
			r.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "EOF", string(body))
		} else {
			fmt.Println("Didn't expect so many requests")
			t.Fatal("Didn't expect so many requests")
		}
		count++
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	// Just run it and see what breaks
	cmd := exec.Command("/bin/bash", "../../build/package/clay-scraper/run.sh", "run-name", "output.txt")
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
}
