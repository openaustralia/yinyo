package test

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/cheggaaa/pb/v3"
	"github.com/openaustralia/yinyo/pkg/apiclient"
	"github.com/openaustralia/yinyo/pkg/protocol"
	"github.com/stretchr/testify/assert"
)

func defaultClient() *apiclient.Client {
	return apiclient.New("http://localhost:8080")
}

func TestHello(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode.")
	}

	client := defaultClient()
	text, err := client.Hello()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "Hello from Yinyo!\n", text)
}

func TestCreateRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode.")
	}

	client := defaultClient()
	run, err := client.CreateRun("foo")
	if err != nil {
		t.Fatal(err)
	}
	defer run.Delete()

	// The only purpose of name_prefix is to make runs easier for humans to identify
	// So, expect the run to start with the name_prefix but there's probably more
	assert.True(t, strings.HasPrefix(run.GetName(), "foo-"))
	assert.NotEqual(t, "", run.GetToken())
}

func TestCreateRunScraperNameEncoding(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode.")
	}

	client := defaultClient()
	run, err := client.CreateRun("foo/b_12r")
	if err != nil {
		t.Fatal(err)
	}
	defer run.Delete()

	// Only certain characters are allowed in kubernetes job names
	assert.True(t, strings.HasPrefix(run.GetName(), "foo-b-12r-"))
}

// Check that run names are created to be unique even when the same scraper name
// is given twice
func TestCreateRunNamesUnique(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode.")
	}

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
	assert.NotEqual(t, run1.GetName(), run2.GetName())
}

func TestNamePrefixOptional(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode.")
	}

	client := defaultClient()
	run, err := client.CreateRun("")
	if err != nil {
		t.Fatal(err)
	}
	defer run.Delete()
	assert.True(t, strings.HasPrefix(run.GetName(), "run-"))
}

func TestUploadDownloadApp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode.")
	}

	// First we need to create a run
	client := defaultClient()
	run, err := client.CreateRun("")
	if err != nil {
		t.Fatal(err)
	}
	defer run.Delete()
	// Now upload an empty tar file (doing this so it validates)
	empty, err := os.Open("fixtures/empty.tgz")
	if err != nil {
		t.Fatal(err)
	}
	err = run.PutApp(empty)
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
	empty2, err := os.Open("fixtures/empty.tgz")
	if err != nil {
		t.Fatal(err)
	}
	b2, err := ioutil.ReadAll(empty2)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, b2, b)
	// TODO: Clean up run
}

// TODO: Add a test for calling CreateRun("TestHelloWorld")

func runScraper(name string, env []protocol.EnvVariable) ([]protocol.Event, error) {
	var eventsList []protocol.Event

	client := defaultClient()
	// Create the run
	run, err := client.CreateRun(name)
	if err != nil {
		return eventsList, err
	}
	defer run.Delete()

	// Now upload the application
	err = run.PutAppFromDirectory("fixtures/scrapers/"+name, []string{})
	if err != nil {
		return eventsList, err
	}

	// Upload the cache if it exists
	file, err := os.Open("fixtures/caches/" + name + ".tar.gz")
	if err == nil {
		err = run.PutCache(file)
		if err != nil {
			return eventsList, err
		}
	} else if !os.IsNotExist(err) {
		return eventsList, err
	}
	// Now start the scraper
	err = run.Start(&protocol.StartRunOptions{Output: "output.txt", Env: env})
	if err != nil {
		return eventsList, err
	}

	// Get the logs (events)
	iterator, err := run.GetEvents("")
	if err != nil {
		return eventsList, err
	}

	// Expect roughly 13 events
	bar := pb.StartNew(13)
	for iterator.More() {
		event, err := iterator.Next()
		if err != nil {
			return eventsList, err
		}
		eventsList = append(eventsList, event)
		bar.Increment()
	}
	bar.Finish()

	// Get the cache
	cache, err := run.GetCache()
	if err != nil {
		return eventsList, err
	}

	file, err = os.Create("fixtures/caches/" + name + ".tar.gz")
	if err != nil {
		return eventsList, err
	}
	_, err = io.Copy(file, cache)
	if err != nil {
		return eventsList, err
	}

	// Get the output
	// Get the metrics
	// Get the exit status
	// Cleanup
	return eventsList, nil
}

func TestHelloWorld(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode.")
	}

	name := "hello-world"
	env := []protocol.EnvVariable{{Name: "HELLO", Value: "Hello World!"}}

	eventsList, err := runScraper(name, env)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: If the cache doesn't exist the test will fail on its first run
	// because the log output is slightly different. Handle this better.

	// Test the running of a super-simple program end-to-end
	// Copy across the IDs and times from the eventsList to the expected because we don't know what they
	// will be ahead of time and this make it easy to compare expected and eventsList
	expected := []protocol.Event{
		protocol.NewStartEvent(eventsList[0].ID, eventsList[0].Time, "build"),
		protocol.NewLogEvent(eventsList[1].ID, eventsList[1].Time, "build", "stdout", "\u001b[1G       \u001b[1G-----> Python app detected"),
		protocol.NewLogEvent(eventsList[2].ID, eventsList[2].Time, "build", "stdout", "\u001b[1G       !     Python has released a security update! Please consider upgrading to python-2.7.16"),
		protocol.NewLogEvent(eventsList[3].ID, eventsList[3].Time, "build", "stdout", "\u001b[1G       Learn More: https://devcenter.heroku.com/articles/python-runtimes"),
		protocol.NewLogEvent(eventsList[4].ID, eventsList[4].Time, "build", "stdout", "\u001b[1G-----> Installing requirements with pip"),
		protocol.NewLogEvent(eventsList[5].ID, eventsList[5].Time, "build", "stdout", "\u001b[1G       You must give at least one requirement to install (see \"pip help install\")"),
		protocol.NewLogEvent(eventsList[6].ID, eventsList[6].Time, "build", "stdout", "\u001b[1G       "),
		protocol.NewLogEvent(eventsList[7].ID, eventsList[7].Time, "build", "stdout", "\u001b[1G       \u001b[1G-----> Discovering process types"),
		protocol.NewLogEvent(eventsList[8].ID, eventsList[8].Time, "build", "stdout", "\u001b[1G       Procfile declares types -> scraper"),
		protocol.NewFinishEvent(eventsList[9].ID, eventsList[9].Time, "build", eventsList[9].Data.(protocol.FinishData).ExitData),
		protocol.NewStartEvent(eventsList[10].ID, eventsList[10].Time, "run"),
		protocol.NewLogEvent(eventsList[11].ID, eventsList[11].Time, "run", "stdout", "Hello World!"),
		protocol.NewFinishEvent(eventsList[12].ID, eventsList[12].Time, "run", eventsList[12].Data.(protocol.FinishData).ExitData),
		protocol.NewLastEvent(eventsList[13].ID, eventsList[13].Time),
	}
	assert.Equal(t, expected, eventsList)
}
