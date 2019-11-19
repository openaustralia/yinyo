package test

// This tests the "clay wrapper" executable without running it in a kubernetes cluster

import (
	"encoding/json"
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

func checkRequestNoBody(t *testing.T, r *http.Request, method string, path string) {
	assert.Equal(t, method, r.Method)
	assert.Equal(t, path, r.URL.EscapedPath())
}

func createTemporaryDirectories() (appPath string, importPath string, cachePath string, err error) {
	currentPath, err := os.Getwd()
	if err != nil {
		return
	}
	appPath, err = ioutil.TempDir(currentPath, "app")
	if err != nil {
		return
	}
	importPath, err = ioutil.TempDir(currentPath, "import")
	if err != nil {
		return
	}
	cachePath, err = ioutil.TempDir(currentPath, "cache")
	return
}

func TestSimpleRun(t *testing.T) {
	count := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		if count == 0 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				`{"stage":"build","type":"start"}`,
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
			// We'll just return the contents of an "arbitrary" directory here. It doesn't
			// really matters what it has in it as long as we can test that it's correct.
			w.Header().Set("Content-Type", "application/gzip")
			reader, err := clayclient.CreateArchiveFromDirectory("fixtures/scrapers/hello-world")
			if err != nil {
				t.Fatal(err)
			}
			_, err = io.Copy(w, reader)
			if err != nil {
				t.Fatal(err)
			}
		} else if count == 3 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				`{"stage":"build","type":"log","stream":"stdout","text":"_app_"}`,
			)
		} else if count == 4 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				`{"stage":"build","type":"log","stream":"stdout","text":"Procfile"}`,
			)
		} else if count == 5 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				`{"stage":"build","type":"log","stream":"stdout","text":"requirements.txt"}`,
			)
		} else if count == 6 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				`{"stage":"build","type":"log","stream":"stdout","text":"runtime.txt"}`,
			)
		} else if count == 7 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				`{"stage":"build","type":"log","stream":"stdout","text":"scraper.py"}`,
			)
		} else if count == 8 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				`{"stage":"build","type":"log","stream":"stdout","text":"_cache_"}`,
			)
		} else if count == 9 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				`{"stage":"build","type":"log","stream":"stdout","text":"requirements.txt"}`,
			)
		} else if count == 10 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				`{"stage":"build","type":"log","stream":"stdout","text":"runtime.txt"}`,
			)
		} else if count == 11 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				`{"stage":"build","type":"log","stream":"stdout","text":"scraper.py"}`,
			)
		} else if count == 12 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				`{"stage":"build","type":"finish"}`,
			)
		} else if count == 13 {
			// We're not testing that the correct thing is being uploaded here for the time being
			checkRequestNoBody(t, r, "PUT", "/runs/run-name/cache")
		} else if count == 14 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				`{"stage":"run","type":"start"}`,
			)
		} else if count == 15 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				`{"stage":"run","type":"log","stream":"stdout","text":"Ran"}`,
			)
		} else if count == 16 {
			checkRequestNoBody(t, r, "PUT", "/runs/run-name/exit-data")
			decoder := json.NewDecoder(r.Body)
			var exitData clayclient.ExitData
			err := decoder.Decode(&exitData)
			if err != nil {
				t.Fatal(err)
			}
			// Check that the exit codes are something sensible
			assert.Equal(t, 0, exitData.Build.ExitCode)
			assert.Equal(t, 0, exitData.Run.ExitCode)
			// The usage values are going to be a little different each time. So, the best we
			// can do for the moment is just check that they are not zero
			assert.True(t, exitData.Build.Usage.WallTime > 0)
			assert.True(t, exitData.Build.Usage.CPUTime > 0)
			assert.True(t, exitData.Build.Usage.MaxRSS > 0)
			assert.True(t, exitData.Build.Usage.NetworkIn > 0)
			assert.True(t, exitData.Build.Usage.NetworkOut > 0)
			assert.True(t, exitData.Run.Usage.WallTime > 0)
			assert.True(t, exitData.Run.Usage.CPUTime > 0)
			assert.True(t, exitData.Run.Usage.MaxRSS > 0)
			assert.True(t, exitData.Run.Usage.NetworkIn > 0)
			assert.True(t, exitData.Run.Usage.NetworkOut > 0)
		} else if count == 17 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				`{"stage":"run","type":"finish"}`,
			)
		} else if count == 18 {
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

	appPath, importPath, cachePath, err := createTemporaryDirectories()
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(appPath)
	defer os.RemoveAll(importPath)
	defer os.RemoveAll(cachePath)

	// Just run it and see what breaks
	cmd := exec.Command(
		"clay",
		"wrapper",
		"--app", appPath,
		"--import", importPath,
		"--cache", cachePath,
		"run-name",
		"output.txt",
	)
	cmd.Env = append(os.Environ(),
		// Send requests for the clay server to our local test server instead (which we start here)
		"CLAY_INTERNAL_SERVER_URL="+ts.URL,
		`CLAY_INTERNAL_BUILD_COMMAND=bash -c "echo _app_; ls `+importPath+`; echo _cache_; ls `+cachePath+`"`,
		"CLAY_INTERNAL_RUN_COMMAND=echo Ran",
	)

	stdoutStderr, err := cmd.CombinedOutput()
	fmt.Printf("%s\n", stdoutStderr)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Test that output is correctly uploaded
}

func TestFailingBuild(t *testing.T) {
	count := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		if count == 0 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				`{"stage":"build","type":"start"}`,
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
				`{"stage":"build","type":"log","stream":"stderr","text":"bash: failing_command: command not found"}`,
			)
		} else if count == 4 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				`{"stage":"build","type":"finish"}`,
			)
		} else if count == 5 {
			// We're not testing that the correct thing is being uploaded here for the time being
			checkRequestNoBody(t, r, "PUT", "/runs/run-name/cache")
		} else if count == 6 {
			checkRequestNoBody(t, r, "PUT", "/runs/run-name/exit-data")
			decoder := json.NewDecoder(r.Body)
			var exitData clayclient.ExitData
			err := decoder.Decode(&exitData)
			if err != nil {
				t.Fatal(err)
			}
			// Check that the exit codes are something sensible
			assert.Equal(t, 127, exitData.Build.ExitCode)
			assert.Nil(t, exitData.Run)
			// The usage values are going to be a little different each time. So, the best we
			// can do for the moment is just check that they are not zero
			assert.True(t, exitData.Build.Usage.WallTime > 0)
			assert.True(t, exitData.Build.Usage.CPUTime > 0)
			assert.True(t, exitData.Build.Usage.MaxRSS > 0)
			assert.True(t, exitData.Build.Usage.NetworkIn > 0)
			assert.True(t, exitData.Build.Usage.NetworkOut > 0)
		} else if count == 7 {
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

	appPath, importPath, cachePath, err := createTemporaryDirectories()
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(appPath)
	defer os.RemoveAll(importPath)
	defer os.RemoveAll(cachePath)

	// Just run it and see what breaks
	cmd := exec.Command(
		"clay",
		"wrapper",
		"--app", appPath,
		"--import", importPath,
		"--cache", cachePath,
		"run-name",
		"output.txt",
	)
	cmd.Env = append(os.Environ(),
		// Send requests for the clay server to our local test server instead (which we start here)
		"CLAY_INTERNAL_SERVER_URL="+ts.URL,
		`CLAY_INTERNAL_BUILD_COMMAND=bash -c "failing_command"`,
		"CLAY_INTERNAL_RUN_COMMAND=echo Ran",
	)

	stdoutStderr, err := cmd.CombinedOutput()
	fmt.Printf("%s\n", stdoutStderr)
	if err != nil {
		log.Fatal(err)
	}
}

func TestFailingRun(t *testing.T) {
	count := 0i
	handler := func(w http.ResponseWriter, r *http.Request) {
		if count == 0 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				`{"stage":"build","type":"start"}`,
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
				`{"stage":"build","type":"log","stream":"stdout","text":"build"}`,
			)
		} else if count == 4 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				`{"stage":"build","type":"finish"}`,
			)
		} else if count == 5 {
			// We're not testing that the correct thing is being uploaded here for the time being
			checkRequestNoBody(t, r, "PUT", "/runs/run-name/cache")
		} else if count == 6 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				`{"stage":"run","type":"start"}`,
			)
		} else if count == 7 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				`{"stage":"run","type":"log","stream":"stderr","text":"bash: failing_command: command not found"}`,
			)
		} else if count == 8 {
			checkRequestNoBody(t, r, "PUT", "/runs/run-name/exit-data")
			decoder := json.NewDecoder(r.Body)
			var exitData clayclient.ExitData
			err := decoder.Decode(&exitData)
			if err != nil {
				t.Fatal(err)
			}
			// Check that the exit codes are something sensible
			assert.Equal(t, 0, exitData.Build.ExitCode)
			assert.Equal(t, 127, exitData.Run.ExitCode)
			// The usage values are going to be a little different each time. So, the best we
			// can do for the moment is just check that they are not zero
			assert.True(t, exitData.Build.Usage.WallTime > 0)
			assert.True(t, exitData.Build.Usage.CPUTime > 0)
			assert.True(t, exitData.Build.Usage.MaxRSS > 0)
			assert.True(t, exitData.Build.Usage.NetworkIn > 0)
			assert.True(t, exitData.Build.Usage.NetworkOut > 0)
			assert.True(t, exitData.Run.Usage.WallTime > 0)
			assert.True(t, exitData.Run.Usage.CPUTime > 0)
			assert.True(t, exitData.Run.Usage.MaxRSS > 0)
			assert.True(t, exitData.Run.Usage.NetworkIn > 0)
			assert.True(t, exitData.Run.Usage.NetworkOut > 0)
		} else if count == 9 {
			checkRequest(t, r,
				"PUT",
				"/runs/run-name/output",
				"hello\n",
			)
		} else if count == 10 {
			checkRequest(t, r,
				"POST",
				"/runs/run-name/events",
				`{"stage":"run","type":"finish"}`,
			)
		} else if count == 11 {
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

	appPath, importPath, cachePath, err := createTemporaryDirectories()
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(appPath)
	defer os.RemoveAll(importPath)
	defer os.RemoveAll(cachePath)

	// Just run it and see what breaks
	cmd := exec.Command(
		"clay",
		"wrapper",
		"--app", appPath,
		"--import", importPath,
		"--cache", cachePath,
		"run-name",
		"output.txt",
	)
	cmd.Env = append(os.Environ(),
		// Send requests for the clay server to our local test server instead (which we start here)
		"CLAY_INTERNAL_SERVER_URL="+ts.URL,
		`CLAY_INTERNAL_BUILD_COMMAND=bash -c "echo build"`,
		// Send something to the output file then fail
		`CLAY_INTERNAL_RUN_COMMAND=bash -c "cd `+appPath+`; echo hello > output.txt; failing_command"`,
	)

	stdoutStderr, err := cmd.CombinedOutput()
	fmt.Printf("%s\n", stdoutStderr)
	if err != nil {
		log.Fatal(err)
	}
}
