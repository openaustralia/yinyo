package client

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func defaultClient() Client {
	return NewClient("http://localhost:8080")
}

func TestHello(t *testing.T) {
	client := defaultClient()
	text, err := client.Hello()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "Hello from Clay!\n", text)
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
	err = client.PutApp(run, body)
	if err != nil {
		t.Fatal(err)
	}

	// Now download the test pattern and check that it matches
	data, err := client.GetApp(run)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ioutil.ReadAll(data)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, app, string(b))
	// TODO: Clean up run
}

// TODO: Add a test for calling CreateRun("TestHelloWorld")

func TestHelloWorld(t *testing.T) {
	// Test the running of a super-simple program end-to-end
	client := defaultClient()
	// Create the run
	run, err := client.CreateRun("test-hello-world")
	if err != nil {
		t.Fatal(err)
	}

	// Now upload the application
	err = client.PutAppFromDirectory(run, "fixtures/scrapers/hello-world")
	if err != nil {
		t.Fatal(err)
	}

	// Upload the cache if it exists
	// TODO: If the cache doesn't exist the test will fail on its first run
	// because the log output is slightly different. Handle this better.
	file, err := os.Open("fixtures/caches/hello-world.tar.gz")
	if err == nil {
		err = client.PutCache(run, file)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		if !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}

	// Now start the scraper
	_, err = client.StartRunRaw(run, &StartRunOptions{Output: "output.txt"})
	if err != nil {
		t.Fatal(err)
	}

	// Get the logs (events)
	events, err := client.GetEventsRaw(run)
	if err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(events.Body)
	var eventStrings []string
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		eventStrings = append(eventStrings, scanner.Text())
	}
	assert.Equal(t, []string{
		"{\"stage\":\"build\",\"type\":\"started\"}",
		"{\"stage\":\"build\",\"type\":\"log\",\"stream\":\"stdout\",\"log\":\"\\u001b[1G       \\u001b[1G-----> Python app detected\"}",
		"{\"stage\":\"build\",\"type\":\"log\",\"stream\":\"stdout\",\"log\":\"\\u001b[1G       !     Python has released a security update! Please consider upgrading to python-2.7.16\"}",
		"{\"stage\":\"build\",\"type\":\"log\",\"stream\":\"stdout\",\"log\":\"\\u001b[1G       Learn More: https://devcenter.heroku.com/articles/python-runtimes\"}",
		"{\"stage\":\"build\",\"type\":\"log\",\"stream\":\"stdout\",\"log\":\"\\u001b[1G-----> Installing requirements with pip\"}",
		"{\"stage\":\"build\",\"type\":\"log\",\"stream\":\"stdout\",\"log\":\"\\u001b[1G       You must give at least one requirement to install (see \\\"pip help install\\\")\"}",
		"{\"stage\":\"build\",\"type\":\"log\",\"stream\":\"stdout\",\"log\":\"\\u001b[1G       \"}",
		"{\"stage\":\"build\",\"type\":\"log\",\"stream\":\"stdout\",\"log\":\"\\u001b[1G       \\u001b[1G-----> Discovering process types\"}",
		"{\"stage\":\"build\",\"type\":\"log\",\"stream\":\"stdout\",\"log\":\"\\u001b[1G       Procfile declares types -> scraper\"}",
		"{\"stage\":\"build\",\"type\":\"finished\"}",
		"{\"stage\":\"run\",\"type\":\"started\"}",
		"{\"stage\":\"run\",\"type\":\"log\",\"stream\":\"stdout\",\"log\":\"Hello World!\"}",
		"{\"stage\":\"run\",\"type\":\"finished\"}",
	}, eventStrings)

	// Get the cache
	cache, err := client.GetCache(run)
	if err != nil {
		t.Fatal(err)
	}

	file, err = os.Create("fixtures/caches/hello-world.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(file, cache)

	// Get the output
	// Get the metrics
	// Get the exit status
	// Cleanup

}
