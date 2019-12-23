package test

// This tests the "yinyo wrapper" executable without running it in a kubernetes cluster

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

	"github.com/openaustralia/yinyo/pkg/event"
	"github.com/openaustralia/yinyo/pkg/yinyoclient"
	"github.com/stretchr/testify/assert"
)

func checkRequestBody(t *testing.T, r *http.Request, body string) {
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

func checkRequestEvent(t *testing.T, r *http.Request, typeString string, data interface{}) {
	dec := json.NewDecoder(r.Body)
	var e event.Event
	err := dec.Decode(&e)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, typeString, e.Type)
	assert.Equal(t, data, e.Data)
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
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "start", event.StartData{Stage: "build"})
		} else if count == 1 {
			checkRequestNoBody(t, r, "GET", "/runs/run-name/app")
			checkRequestBody(t, r, "")
			w.Header().Set("Content-Type", "application/gzip")
			reader, err := yinyoclient.CreateArchiveFromDirectory("fixtures/scrapers/hello-world", []string{})
			if err != nil {
				t.Fatal(err)
			}
			_, err = io.Copy(w, reader)
			if err != nil {
				t.Fatal(err)
			}
		} else if count == 2 {
			checkRequestNoBody(t, r, "GET", "/runs/run-name/cache")
			checkRequestBody(t, r, "")
			// We'll just return the contents of an "arbitrary" directory here. It doesn't
			// really matters what it has in it as long as we can test that it's correct.
			w.Header().Set("Content-Type", "application/gzip")
			reader, err := yinyoclient.CreateArchiveFromDirectory("fixtures/scrapers/hello-world", []string{})
			if err != nil {
				t.Fatal(err)
			}
			_, err = io.Copy(w, reader)
			if err != nil {
				t.Fatal(err)
			}
		} else if count == 3 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", event.LogData{Stage: "build", Stream: "stdout", Text: "_app_"})
		} else if count == 4 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", event.LogData{Stage: "build", Stream: "stdout", Text: "Procfile"})
		} else if count == 5 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", event.LogData{Stage: "build", Stream: "stdout", Text: "requirements.txt"})
		} else if count == 6 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", event.LogData{Stage: "build", Stream: "stdout", Text: "runtime.txt"})
		} else if count == 7 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", event.LogData{Stage: "build", Stream: "stdout", Text: "scraper.py"})
		} else if count == 8 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", event.LogData{Stage: "build", Stream: "stdout", Text: "_cache_"})
		} else if count == 9 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", event.LogData{Stage: "build", Stream: "stdout", Text: "requirements.txt"})
		} else if count == 10 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", event.LogData{Stage: "build", Stream: "stdout", Text: "runtime.txt"})
		} else if count == 11 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", event.LogData{Stage: "build", Stream: "stdout", Text: "scraper.py"})
		} else if count == 12 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "finish", event.FinishData{Stage: "build"})
		} else if count == 13 {
			// We're not testing that the correct thing is being uploaded here for the time being
			checkRequestNoBody(t, r, "PUT", "/runs/run-name/cache")
		} else if count == 14 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "start", event.StartData{Stage: "run"})
		} else if count == 15 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", event.LogData{Stage: "run", Stream: "stdout", Text: "Ran"})
		} else if count == 16 {
			checkRequestNoBody(t, r, "PUT", "/runs/run-name/exit-data")
			decoder := json.NewDecoder(r.Body)
			var exitData yinyoclient.ExitData
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
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "finish", event.FinishData{Stage: "run"})
		} else if count == 18 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "last", event.LastData{})
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
	// TODO: Don't run the executable
	cmd := exec.Command(
		"yinyo",
		"wrapper",
		"--apppath", appPath,
		"--importpath", importPath,
		"--cachepath", cachePath,
		"run-name",
		"run-token",
		"--output", "output.txt",
		// Send requests for the yinyo server to our local test server instead (which we start here)
		"--server", ts.URL,
		"--buildcommand", `bash -c "echo _app_; ls `+importPath+`; echo _cache_; ls `+cachePath+`"`,
		"--runcommand", "echo Ran",
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
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "start", event.StartData{Stage: "build"})
		} else if count == 1 {
			checkRequestNoBody(t, r, "GET", "/runs/run-name/app")
			checkRequestBody(t, r, "")
			w.Header().Set("Content-Type", "application/gzip")
			reader, err := yinyoclient.CreateArchiveFromDirectory("fixtures/scrapers/hello-world", []string{})
			if err != nil {
				t.Fatal(err)
			}
			_, err = io.Copy(w, reader)
			if err != nil {
				t.Fatal(err)
			}
		} else if count == 2 {
			checkRequestNoBody(t, r, "GET", "/runs/run-name/cache")
			checkRequestBody(t, r, "")
			// Let the client know that there is no cache in this case
			http.NotFound(w, r)
		} else if count == 3 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", event.LogData{Stage: "build", Stream: "stderr", Text: "bash: failing_command: command not found"})
		} else if count == 4 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "finish", event.FinishData{Stage: "build"})
		} else if count == 5 {
			// We're not testing that the correct thing is being uploaded here for the time being
			checkRequestNoBody(t, r, "PUT", "/runs/run-name/cache")
		} else if count == 6 {
			checkRequestNoBody(t, r, "PUT", "/runs/run-name/exit-data")
			decoder := json.NewDecoder(r.Body)
			var exitData yinyoclient.ExitData
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
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "last", event.LastData{})
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
		"yinyo",
		"wrapper",
		"--apppath", appPath,
		"--importpath", importPath,
		"--cachepath", cachePath,
		"run-name",
		"run-token",
		"--output", "output.txt",
		// Send requests for the yinyo server to our local test server instead (which we start here)
		"--server", ts.URL,
		"--buildcommand", `bash -c "failing_command"`,
		"--runcommand", "echo Ran",
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
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "start", event.StartData{Stage: "build"})
		} else if count == 1 {
			checkRequestNoBody(t, r, "GET", "/runs/run-name/app")
			checkRequestBody(t, r, "")
			w.Header().Set("Content-Type", "application/gzip")
			reader, err := yinyoclient.CreateArchiveFromDirectory("fixtures/scrapers/hello-world", []string{})
			if err != nil {
				t.Fatal(err)
			}
			_, err = io.Copy(w, reader)
			if err != nil {
				t.Fatal(err)
			}
		} else if count == 2 {
			checkRequestNoBody(t, r, "GET", "/runs/run-name/cache")
			checkRequestBody(t, r, "")
			// Let the client know that there is no cache in this case
			http.NotFound(w, r)
		} else if count == 3 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", event.LogData{Stage: "build", Stream: "stdout", Text: "build"})
		} else if count == 4 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "finish", event.FinishData{Stage: "build"})
		} else if count == 5 {
			// We're not testing that the correct thing is being uploaded here for the time being
			checkRequestNoBody(t, r, "PUT", "/runs/run-name/cache")
		} else if count == 6 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "start", event.StartData{Stage: "run"})
		} else if count == 7 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", event.LogData{Stage: "run", Stream: "stderr", Text: "bash: failing_command: command not found"})
		} else if count == 8 {
			checkRequestNoBody(t, r, "PUT", "/runs/run-name/exit-data")
			decoder := json.NewDecoder(r.Body)
			var exitData yinyoclient.ExitData
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
			checkRequestNoBody(t, r, "PUT", "/runs/run-name/output")
			checkRequestBody(t, r, "hello\n")
		} else if count == 10 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "finish", event.FinishData{Stage: "run"})
		} else if count == 11 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "last", event.LastData{})
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
		"yinyo",
		"wrapper",
		"--apppath", appPath,
		"--importpath", importPath,
		"--cachepath", cachePath,
		"run-name",
		"run-token",
		"--output", "output.txt",
		// Send requests for the yinyo server to our local test server instead (which we start here)
		"--server", ts.URL,
		"--buildcommand", `bash -c "echo build"`,
		// Send something to the output file then fail
		"--runcommand", `bash -c "cd `+appPath+`; echo hello > output.txt; failing_command"`,
	)

	stdoutStderr, err := cmd.CombinedOutput()
	fmt.Printf("%s\n", stdoutStderr)
	if err != nil {
		log.Fatal(err)
	}
}

func TestInternalError(t *testing.T) {
	// If the wrapper has an error, either from doing something itself or from contacting
	// the yinyo server, it should also add something to the log to let the user know
	count := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		if count == 0 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "start", event.StartData{Stage: "build"})
		} else if count == 1 {
			// Let's simulate an error with the blob storage. So, the wrapper is trying to
			// get the application and there's a problem.
			checkRequestNoBody(t, r, "GET", "/runs/run-name/app")
			checkRequestBody(t, r, "")
			w.WriteHeader(http.StatusInternalServerError)
		} else if count == 2 {
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", event.LogData{Stage: "build", Stream: "interr", Text: "Could not get the code"})
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
	// TODO: Don't run the executable
	cmd := exec.Command(
		"yinyo",
		"wrapper",
		"--apppath", appPath,
		"--importpath", importPath,
		"--cachepath", cachePath,
		"run-name",
		"run-token",
		"--output", "output.txt",
		// Send requests for the yinyo server to our local test server instead (which we start here)
		"--server", ts.URL,
		"--buildcommand", "echo Build",
		"--runcommand", "echo Ran",
	)

	stdoutStderr, err := cmd.CombinedOutput()
	fmt.Printf("%s\n", stdoutStderr)

	// Because we expect the command to fail
	assert.NotNil(t, err)
	assert.NotEqual(t, 0, cmd.ProcessState.ExitCode())

	assert.Equal(t, 3, count)
}
