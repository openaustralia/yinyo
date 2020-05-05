package test

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/cheggaaa/pb/v3"
	"github.com/openaustralia/yinyo/pkg/apiclient"
	"github.com/openaustralia/yinyo/pkg/protocol"
	uuid "github.com/satori/go.uuid"
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
	hello, err := client.Hello()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, protocol.DefaultAndMax{Default: 3600, Max: 86400}, hello.MaxRunTime)
	assert.Equal(t, protocol.DefaultAndMax{Default: 1073741824, Max: 1610612736}, hello.Memory)
	assert.Equal(t, protocol.DefaultAndMax{Default: 1073741824, Max: 1610612736}, hello.Memory)
	// We can't say for sure what the git revision is going to be. So, just test for the length
	assert.Equal(t, 40, len(hello.Version))
}

func TestCreateRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode.")
	}

	client := defaultClient()
	run, err := client.CreateRun(protocol.CreateRunOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer run.Delete()

	// Check that run id looks like a uuid
	_, err = uuid.FromString(run.GetID())
	assert.Nil(t, err)
}

func TestUploadDownloadApp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode.")
	}

	// First we need to create a run
	client := defaultClient()
	run, err := client.CreateRun(protocol.CreateRunOptions{})
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

// TODO: Use high-level client library instead?
func runScraper(appDirectory string, cachePath string, env []protocol.EnvVariable) ([]protocol.Event, error) {
	var eventsList []protocol.Event

	client := defaultClient()
	// Create the run
	run, err := client.CreateRun(protocol.CreateRunOptions{})
	if err != nil {
		return eventsList, err
	}
	defer run.Delete()

	// Now upload the application
	err = run.PutAppFromDirectory(appDirectory, []string{})
	if err != nil {
		return eventsList, err
	}

	// Upload the cache if it exists
	file, err := os.Open(cachePath)
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

	file, err = os.Create(cachePath)
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

func runScraperWithPreCache(name string, env []protocol.EnvVariable) ([]protocol.Event, error) {
	appDirectory := "fixtures/scrapers/" + name
	cachePath := "fixtures/caches/" + name + ".tar.gz"

	var eventsList []protocol.Event
	// Check if cache doesn't exist in which case we'll generate it by running
	// the whole scraper before we run the main tests
	_, err := os.Stat(cachePath)
	if os.IsNotExist(err) {
		_, err = runScraper(appDirectory, cachePath, env)
		if err != nil {
			return eventsList, err
		}
	}
	eventsList, err = runScraper(appDirectory, cachePath, env)
	if err != nil {
		return eventsList, err
	}
	return eventsList, nil
}

func TestHelloWorld(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode.")
	}

	eventsList, err := runScraperWithPreCache("hello-world", []protocol.EnvVariable{{Name: "HELLO", Value: "Hello World!"}})
	if err != nil {
		log.Fatal(err)
	}

	// Test the running of a super-simple program end-to-end
	// Copy across the IDs and times from the eventsList to the expected because we don't know what they
	// will be ahead of time and this make it easy to compare expected and eventsList
	expected := []protocol.Event{
		protocol.NewFirstEvent(eventsList[0].ID, eventsList[0].RunID, eventsList[0].Time),
		protocol.NewStartEvent(eventsList[1].ID, eventsList[1].RunID, eventsList[1].Time, "build"),
		protocol.NewLogEvent(eventsList[2].ID, eventsList[2].RunID, eventsList[2].Time, "build", "stdout", "\u001b[1G       \u001b[1G-----> Python app detected"),
		protocol.NewLogEvent(eventsList[3].ID, eventsList[3].RunID, eventsList[3].Time, "build", "stdout", "\u001b[1G       !     Python has released a security update! Please consider upgrading to python-2.7.16"),
		protocol.NewLogEvent(eventsList[4].ID, eventsList[4].RunID, eventsList[4].Time, "build", "stdout", "\u001b[1G       Learn More: https://devcenter.heroku.com/articles/python-runtimes"),
		protocol.NewLogEvent(eventsList[5].ID, eventsList[5].RunID, eventsList[5].Time, "build", "stdout", "\u001b[1G-----> Installing requirements with pip"),
		protocol.NewLogEvent(eventsList[6].ID, eventsList[6].RunID, eventsList[6].Time, "build", "stdout", "\u001b[1G       You must give at least one requirement to install (see \"pip help install\")"),
		protocol.NewLogEvent(eventsList[7].ID, eventsList[7].RunID, eventsList[7].Time, "build", "stdout", "\u001b[1G       "),
		protocol.NewLogEvent(eventsList[8].ID, eventsList[8].RunID, eventsList[8].Time, "build", "stdout", "\u001b[1G       \u001b[1G-----> Discovering process types"),
		protocol.NewLogEvent(eventsList[9].ID, eventsList[9].RunID, eventsList[9].Time, "build", "stdout", "\u001b[1G       Procfile declares types -> scraper"),
		protocol.NewFinishEvent(eventsList[10].ID, eventsList[10].RunID, eventsList[10].Time, "build", eventsList[10].Data.(protocol.FinishData).ExitData),
		protocol.NewStartEvent(eventsList[11].ID, eventsList[11].RunID, eventsList[11].Time, "run"),
		protocol.NewLogEvent(eventsList[12].ID, eventsList[12].RunID, eventsList[12].Time, "run", "stdout", "Hello World!"),
		protocol.NewFinishEvent(eventsList[13].ID, eventsList[13].RunID, eventsList[13].Time, "run", eventsList[13].Data.(protocol.FinishData).ExitData),
		protocol.NewLastEvent(eventsList[14].ID, eventsList[14].RunID, eventsList[14].Time),
	}
	assert.Equal(t, expected, eventsList)
}
