package test

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/cheggaaa/pb/v3"
	"github.com/openaustralia/yinyo/pkg/event"
	"github.com/openaustralia/yinyo/pkg/yinyoclient"
	"github.com/stretchr/testify/assert"
)

func defaultClient() *yinyoclient.Client {
	return yinyoclient.New("http://localhost:8080")
}

func TestHello(t *testing.T) {
	client := defaultClient()
	text, err := client.Hello()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "Hello from Yinyo!\n", text)
}

func TestCreateRun(t *testing.T) {
	client := defaultClient()
	run, err := client.CreateRun("foo")
	if err != nil {
		t.Fatal(err)
	}
	defer run.Delete()

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
	defer run.Delete()

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
	defer run1.Delete()
	run2, err := client.CreateRun("foo")
	if err != nil {
		t.Fatal(err)
	}
	defer run2.Delete()
	assert.NotEqual(t, run1.Name, run2.Name)
}

func TestNamePrefixOptional(t *testing.T) {
	client := defaultClient()
	run, err := client.CreateRun("")
	if err != nil {
		t.Fatal(err)
	}
	defer run.Delete()
	assert.True(t, strings.HasPrefix(run.Name, "run-"))
}

func TestUploadDownloadApp(t *testing.T) {
	// First we need to create a run
	client := defaultClient()
	run, err := client.CreateRun("")
	if err != nil {
		t.Fatal(err)
	}
	defer run.Delete()
	// Now upload a random test pattern for the app
	app := "Random test pattern"
	body := strings.NewReader(app)
	err = run.PutApp(body)
	if err != nil {
		t.Fatal(err)
	}

	// Now download the test pattern and check that it matches
	data, err := run.GetApp()
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
	defer run.Delete()

	// Now upload the application
	err = run.PutAppFromDirectory("fixtures/scrapers/hello-world", []string{})
	if err != nil {
		t.Fatal(err)
	}

	// Upload the cache if it exists
	// TODO: If the cache doesn't exist the test will fail on its first run
	// because the log output is slightly different. Handle this better.
	file, err := os.Open("fixtures/caches/hello-world.tar.gz")
	if err == nil {
		err = run.PutCache(file)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		if !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}

	// Now start the scraper
	err = run.Start(&yinyoclient.StartRunOptions{Output: "output.txt", Env: []yinyoclient.EnvVariable{
		yinyoclient.EnvVariable{Name: "HELLO", Value: "Hello World!"},
	}})
	if err != nil {
		t.Fatal(err)
	}

	// Get the logs (events)
	iterator, err := run.GetEvents("")
	if err != nil {
		t.Fatal(err)
	}

	var eventsList []event.Event
	// Expect roughly 13 events
	bar := pb.StartNew(13)
	for iterator.More() {
		event, err := iterator.Next()
		if err != nil {
			t.Fatal(err)
		}
		eventsList = append(eventsList, event)
		bar.Increment()
	}
	bar.Finish()
	// Copy across the IDs from the eventsList to the expected because we don't know what the
	// IDs will be ahead of time and this make it easy to compare expected and eventsList
	expected := []event.Event{
		event.NewStartEvent(eventsList[0].ID, "build"),
		event.NewLogEvent(eventsList[1].ID, "build", "stdout", "\u001b[1G       \u001b[1G-----> Python app detected"),
		event.NewLogEvent(eventsList[2].ID, "build", "stdout", "\u001b[1G       !     Python has released a security update! Please consider upgrading to python-2.7.16"),
		event.NewLogEvent(eventsList[3].ID, "build", "stdout", "\u001b[1G       Learn More: https://devcenter.heroku.com/articles/python-runtimes"),
		event.NewLogEvent(eventsList[4].ID, "build", "stdout", "\u001b[1G-----> Installing requirements with pip"),
		event.NewLogEvent(eventsList[5].ID, "build", "stdout", "\u001b[1G       You must give at least one requirement to install (see \"pip help install\")"),
		event.NewLogEvent(eventsList[6].ID, "build", "stdout", "\u001b[1G       "),
		event.NewLogEvent(eventsList[7].ID, "build", "stdout", "\u001b[1G       \u001b[1G-----> Discovering process types"),
		event.NewLogEvent(eventsList[8].ID, "build", "stdout", "\u001b[1G       Procfile declares types -> scraper"),
		event.NewFinishEvent(eventsList[9].ID, "build"),
		event.NewStartEvent(eventsList[10].ID, "run"),
		event.NewLogEvent(eventsList[11].ID, "run", "stdout", "Hello World!"),
		event.NewFinishEvent(eventsList[12].ID, "run"),
		event.NewLastEvent(eventsList[13].ID),
	}
	assert.Equal(t, expected, eventsList)

	// Get the cache
	cache, err := run.GetCache()
	if err != nil {
		t.Fatal(err)
	}

	file, err = os.Create("fixtures/caches/hello-world.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(file, cache)
	if err != nil {
		t.Fatal(err)
	}

	// Get the output
	// Get the metrics
	// Get the exit status
	// Cleanup
}
