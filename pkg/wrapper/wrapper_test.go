package wrapper

// This tests the "yinyo wrapper" without running it in a kubernetes cluster

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/openaustralia/yinyo/pkg/archive"
	"github.com/openaustralia/yinyo/pkg/protocol"
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
	var e protocol.Event
	err := dec.Decode(&e)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, typeString, e.Type)
	assert.Equal(t, data, e.Data)
}

func createTemporaryDirectories() (appPath string, importPath string, cachePath string, envPath string, err error) {
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
	if err != nil {
		return
	}
	envPath, err = ioutil.TempDir(currentPath, "env")
	return
}

func TestSimpleRun(t *testing.T) {
	count := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch count {
		case 0:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "start", protocol.StartData{Stage: "build"})
		case 1:
			checkRequestNoBody(t, r, "GET", "/runs/run-name/app")
			checkRequestBody(t, r, "")
			w.Header().Set("Content-Type", "application/gzip")
			reader, err := archive.CreateFromDirectory("fixtures/scrapers/hello-world", []string{})
			if err != nil {
				t.Fatal(err)
			}
			_, err = io.Copy(w, reader)
			if err != nil {
				t.Fatal(err)
			}
		case 2:
			checkRequestNoBody(t, r, "GET", "/runs/run-name/cache")
			checkRequestBody(t, r, "")
			// We'll just return the contents of an "arbitrary" directory here. It doesn't
			// really matters what it has in it as long as we can test that it's correct.
			w.Header().Set("Content-Type", "application/gzip")
			reader, err := archive.CreateFromDirectory("fixtures/scrapers/hello-world", []string{})
			if err != nil {
				t.Fatal(err)
			}
			_, err = io.Copy(w, reader)
			if err != nil {
				t.Fatal(err)
			}
		case 3:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", protocol.LogData{Stage: "build", Stream: "stdout", Text: "_app_"})
		case 4:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", protocol.LogData{Stage: "build", Stream: "stdout", Text: "Procfile"})
		case 5:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", protocol.LogData{Stage: "build", Stream: "stdout", Text: "requirements.txt"})
		case 6:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", protocol.LogData{Stage: "build", Stream: "stdout", Text: "runtime.txt"})
		case 7:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", protocol.LogData{Stage: "build", Stream: "stdout", Text: "scraper.py"})
		case 8:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", protocol.LogData{Stage: "build", Stream: "stdout", Text: "_cache_"})
		case 9:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", protocol.LogData{Stage: "build", Stream: "stdout", Text: "requirements.txt"})
		case 10:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", protocol.LogData{Stage: "build", Stream: "stdout", Text: "runtime.txt"})
		case 11:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", protocol.LogData{Stage: "build", Stream: "stdout", Text: "scraper.py"})
		case 12:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "finish", protocol.FinishData{Stage: "build"})
		case 13:
			// We're not testing that the correct thing is being uploaded here for the time being
			checkRequestNoBody(t, r, "PUT", "/runs/run-name/cache")
		case 14:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "start", protocol.StartData{Stage: "run"})
		case 15:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", protocol.LogData{Stage: "run", Stream: "stdout", Text: "Ran"})
		case 16:
			checkRequestNoBody(t, r, "PUT", "/runs/run-name/exit-data")
			decoder := json.NewDecoder(r.Body)
			var exitData protocol.ExitData
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
		case 17:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "finish", protocol.FinishData{Stage: "run"})
		case 18:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "last", protocol.LastData{})
		default:
			fmt.Println("Didn't expect so many requests")
			t.Fatal("Didn't expect so many requests")
		}
		count++
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	appPath, importPath, cachePath, envPath, err := createTemporaryDirectories()
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(appPath)
	defer os.RemoveAll(importPath)
	defer os.RemoveAll(cachePath)
	defer os.RemoveAll(envPath)

	err = Run(Options{
		RunName:  "run-name",
		RunToken: "run-token",
		// Send requests for the yinyo server to our local test server instead (which we start here)
		ServerURL:    ts.URL,
		ImportPath:   importPath,
		CachePath:    cachePath,
		AppPath:      appPath,
		EnvPath:      envPath,
		BuildCommand: `bash -c "echo _app_; ls ` + importPath + `; echo _cache_; ls ` + cachePath + `"`,
		RunCommand:   "echo Ran",
		RunOutput:    "output.txt",
	})
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Test that output is correctly uploaded
}

func TestFailingBuild(t *testing.T) {
	count := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch count {
		case 0:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "start", protocol.StartData{Stage: "build"})
		case 1:
			checkRequestNoBody(t, r, "GET", "/runs/run-name/app")
			checkRequestBody(t, r, "")
			w.Header().Set("Content-Type", "application/gzip")
			reader, err := archive.CreateFromDirectory("fixtures/scrapers/hello-world", []string{})
			if err != nil {
				t.Fatal(err)
			}
			_, err = io.Copy(w, reader)
			if err != nil {
				t.Fatal(err)
			}
		case 2:
			checkRequestNoBody(t, r, "GET", "/runs/run-name/cache")
			checkRequestBody(t, r, "")
			// Let the client know that there is no cache in this case
			http.NotFound(w, r)
		case 3:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", protocol.LogData{Stage: "build", Stream: "stderr", Text: "bash: failing_command: command not found"})
		case 4:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "finish", protocol.FinishData{Stage: "build"})
		case 5:
			// We're not testing that the correct thing is being uploaded here for the time being
			checkRequestNoBody(t, r, "PUT", "/runs/run-name/cache")
		case 6:
			checkRequestNoBody(t, r, "PUT", "/runs/run-name/exit-data")
			decoder := json.NewDecoder(r.Body)
			var exitData protocol.ExitData
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
		case 7:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "last", protocol.LastData{})
		default:
			fmt.Println("Didn't expect so many requests")
			t.Fatal("Didn't expect so many requests")
		}
		count++
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	appPath, importPath, cachePath, envPath, err := createTemporaryDirectories()
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(appPath)
	defer os.RemoveAll(importPath)
	defer os.RemoveAll(cachePath)
	defer os.RemoveAll(envPath)

	err = Run(Options{
		RunName:  "run-name",
		RunToken: "run-token",
		// Send requests for the yinyo server to our local test server instead (which we start here)
		ServerURL:    ts.URL,
		ImportPath:   importPath,
		CachePath:    cachePath,
		AppPath:      appPath,
		EnvPath:      envPath,
		BuildCommand: `bash -c "failing_command"`,
		RunCommand:   "echo Ran",
		RunOutput:    "output.txt",
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestFailingRun(t *testing.T) {
	count := 0i
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch count {
		case 0:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "start", protocol.StartData{Stage: "build"})
		case 1:
			checkRequestNoBody(t, r, "GET", "/runs/run-name/app")
			checkRequestBody(t, r, "")
			w.Header().Set("Content-Type", "application/gzip")
			reader, err := archive.CreateFromDirectory("fixtures/scrapers/hello-world", []string{})
			if err != nil {
				t.Fatal(err)
			}
			_, err = io.Copy(w, reader)
			if err != nil {
				t.Fatal(err)
			}
		case 2:
			checkRequestNoBody(t, r, "GET", "/runs/run-name/cache")
			checkRequestBody(t, r, "")
			// Let the client know that there is no cache in this case
			http.NotFound(w, r)
		case 3:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", protocol.LogData{Stage: "build", Stream: "stdout", Text: "build"})
		case 4:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "finish", protocol.FinishData{Stage: "build"})
		case 5:
			// We're not testing that the correct thing is being uploaded here for the time being
			checkRequestNoBody(t, r, "PUT", "/runs/run-name/cache")
		case 6:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "start", protocol.StartData{Stage: "run"})
		case 7:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", protocol.LogData{Stage: "run", Stream: "stderr", Text: "bash: failing_command: command not found"})
		case 8:
			checkRequestNoBody(t, r, "PUT", "/runs/run-name/exit-data")
			decoder := json.NewDecoder(r.Body)
			var exitData protocol.ExitData
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
		case 9:
			checkRequestNoBody(t, r, "PUT", "/runs/run-name/output")
			checkRequestBody(t, r, "hello\n")
		case 10:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "finish", protocol.FinishData{Stage: "run"})
		case 11:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "last", protocol.LastData{})
		default:
			fmt.Println("Didn't expect so many requests")
			t.Fatal("Didn't expect so many requests")
		}
		count++
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	appPath, importPath, cachePath, envPath, err := createTemporaryDirectories()
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(appPath)
	defer os.RemoveAll(importPath)
	defer os.RemoveAll(cachePath)
	defer os.RemoveAll(envPath)

	err = Run(Options{
		RunName:  "run-name",
		RunToken: "run-token",
		// Send requests for the yinyo server to our local test server instead (which we start here)
		ServerURL:    ts.URL,
		ImportPath:   importPath,
		CachePath:    cachePath,
		AppPath:      appPath,
		EnvPath:      envPath,
		BuildCommand: `bash -c "echo build"`,
		// Send something to the output file then fail
		RunCommand: `bash -c "cd ` + appPath + `; echo hello > output.txt; failing_command"`,
		RunOutput:  "output.txt",
	})
	if err != nil {
		log.Fatal(err)
	}
}

func TestInternalError(t *testing.T) {
	// If the wrapper has an error, either from doing something itself or from contacting
	// the yinyo server, it should also add something to the log to let the user know
	count := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch count {
		case 0:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "start", protocol.StartData{Stage: "build"})
		case 1:
			// Let's simulate an error with the blob storage. So, the wrapper is trying to
			// get the application and there's a problem.
			checkRequestNoBody(t, r, "GET", "/runs/run-name/app")
			checkRequestBody(t, r, "")
			w.WriteHeader(http.StatusInternalServerError)
		case 2:
			checkRequestNoBody(t, r, "POST", "/runs/run-name/events")
			checkRequestEvent(t, r, "log", protocol.LogData{Stage: "", Stream: "interr", Text: "Internal error"})
		}
		count++
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	appPath, importPath, cachePath, envPath, err := createTemporaryDirectories()
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(appPath)
	defer os.RemoveAll(importPath)
	defer os.RemoveAll(cachePath)
	defer os.RemoveAll(envPath)

	err = Run(Options{
		RunName:  "run-name",
		RunToken: "run-token",
		// Send requests for the yinyo server to our local test server instead (which we start here)
		ServerURL:    ts.URL,
		ImportPath:   importPath,
		CachePath:    cachePath,
		AppPath:      appPath,
		EnvPath:      envPath,
		BuildCommand: "echo Build",
		RunCommand:   "echo Ran",
		RunOutput:    "output.txt",
	})
	// Because we expect the command to fail
	assert.NotNil(t, err)
	assert.Equal(t, 3, count)
}
