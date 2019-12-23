package commands

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/openaustralia/yinyo/pkg/blobstore"
	"github.com/openaustralia/yinyo/pkg/event"
	"github.com/openaustralia/yinyo/pkg/jobdispatcher"
	"github.com/openaustralia/yinyo/pkg/keyvaluestore"
	"github.com/openaustralia/yinyo/pkg/stream"
)

func TestStoragePath(t *testing.T) {
	assert.Equal(t, storagePath("abc", "app.tgz"), "abc/app.tgz")
	assert.Equal(t, storagePath("def", "output"), "def/output")
}

func TestStartRun(t *testing.T) {
	job := new(jobdispatcher.MockClient)
	keyValueStore := new(keyvaluestore.MockClient)
	blobStore := new(blobstore.MockClient)

	// Expect that the job will get dispatched
	job.On(
		"StartJob",
		"run-name",
		"openaustralia/yinyo-scraper:v1",
		[]string{"/bin/yinyo", "wrapper", "run-name", "supersecret", "--output", "output.txt", "--env", "FOO=bar"},
	).Return(nil)
	// Expect that we'll need the secret token
	job.On("GetToken", "run-name").Return("supersecret", nil)
	// Expect that we save the callback url in the key value store
	keyValueStore.On("Set", "url:run-name", "http://foo.com").Return(nil)
	// Expect that we try to get the code just to see if it exists
	blobStore.On("Get", "run-name/app.tgz").Return(nil, nil)

	app := AppImplementation{JobDispatcher: job, KeyValueStore: keyValueStore, BlobStore: blobStore}
	// TODO: Pass an options struct instead (we get named parameters effectively then)
	err := app.StartRun(
		"run-name",                      // Run name
		"output.txt",                    // Output filename
		map[string]string{"FOO": "bar"}, // Environment variables
		"http://foo.com",                // Callback URL
	)
	assert.Nil(t, err)

	job.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
	blobStore.AssertExpectations(t)
}

type MockRoundTripper struct {
	mock.Mock
}

func (m *MockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	args := m.Called(r)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestCreateEvent(t *testing.T) {
	stream := new(stream.MockClient)
	keyValueStore := new(keyvaluestore.MockClient)

	time := time.Now()
	stream.On("Add", "run-name", event.NewStartEvent("", time, "build")).Return(event.NewStartEvent("123", time, "build"), nil)
	keyValueStore.On("Get", "url:run-name").Return("http://foo.com/bar", nil)

	// Mock out the http RoundTripper so that no actual http request is made
	httpClient := defaultHTTP()
	roundTripper := new(MockRoundTripper)
	roundTripper.On("RoundTrip", mock.MatchedBy(func(r *http.Request) bool {
		return r.URL.String() == "http://foo.com/bar"
	})).Return(
		&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader("")),
		},
		nil,
	)
	httpClient.Transport = roundTripper

	app := AppImplementation{Stream: stream, KeyValueStore: keyValueStore, HTTP: httpClient}
	err := app.CreateEvent("run-name", event.NewStartEvent("", time, "build"))
	assert.Nil(t, err)

	stream.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
	roundTripper.AssertExpectations(t)
}

func TestCreateEventNoCallbackURL(t *testing.T) {
	stream := new(stream.MockClient)
	keyValueStore := new(keyvaluestore.MockClient)

	time := time.Now()
	stream.On("Add", "run-name", event.NewStartEvent("", time, "build")).Return(event.NewStartEvent("123", time, "build"), nil)
	keyValueStore.On("Get", "url:run-name").Return("", nil)

	// Mock out the http RoundTripper so that no actual http request is made
	httpClient := defaultHTTP()
	roundTripper := new(MockRoundTripper)
	httpClient.Transport = roundTripper

	app := AppImplementation{Stream: stream, KeyValueStore: keyValueStore, HTTP: httpClient}
	err := app.CreateEvent("run-name", event.NewStartEvent("", time, "build"))
	assert.Nil(t, err)

	stream.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
	roundTripper.AssertNotCalled(t, "RoundTrip")
}

func TestCreateEventErrorDuringCallback(t *testing.T) {
	stream := new(stream.MockClient)
	keyValueStore := new(keyvaluestore.MockClient)

	time := time.Now()
	stream.On("Add", "run-name", event.NewStartEvent("", time, "build")).Return(event.NewStartEvent("123", time, "build"), nil)
	keyValueStore.On("Get", "url:run-name").Return("http://foo.com/bar", nil)

	// Mock out the http RoundTripper so that no actual http request is made
	httpClient := defaultHTTP()
	roundTripper := new(MockRoundTripper)
	roundTripper.On("RoundTrip", mock.MatchedBy(func(r *http.Request) bool {
		return r.URL.String() == "http://foo.com/bar"
	})).Return(
		&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader("")),
		},
		errors.New("An error while doing the postback"),
	)
	httpClient.Transport = roundTripper

	app := AppImplementation{Stream: stream, KeyValueStore: keyValueStore, HTTP: httpClient}
	err := app.CreateEvent("run-name", event.NewStartEvent("", time, "build"))
	assert.EqualError(t, err, "Post http://foo.com/bar: An error while doing the postback")

	stream.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
	roundTripper.AssertExpectations(t)
}

func TestGetEvents(t *testing.T) {
	stream := new(stream.MockClient)

	time := time.Now()
	stream.On("Get", "run-name", "0").Return(event.NewStartEvent("123", time, "build"), nil)
	stream.On("Get", "run-name", "123").Return(event.NewLastEvent("456", time), nil)

	app := AppImplementation{Stream: stream}

	events := app.GetEvents("run-name", "0")

	// We're expecting two events in the stream. Let's hardcode what would normally be in a loop
	assert.True(t, events.More())
	e, err := events.Next()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, event.NewStartEvent("123", time, "build"), e)
	assert.True(t, events.More())
	e, err = events.Next()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, event.NewLastEvent("456", time), e)
	assert.False(t, events.More())

	stream.AssertExpectations(t)
}

func TestDeleteRun(t *testing.T) {
	jobDispatcher := new(jobdispatcher.MockClient)
	blobStore := new(blobstore.MockClient)
	stream := new(stream.MockClient)
	keyValueStore := new(keyvaluestore.MockClient)

	jobDispatcher.On("DeleteJobAndToken", "run-name").Return(nil)
	blobStore.On("Delete", "run-name/app.tgz").Return(nil)
	blobStore.On("Delete", "run-name/output").Return(nil)
	blobStore.On("Delete", "run-name/exit-data.json").Return(nil)
	blobStore.On("Delete", "run-name/cache.tgz").Return(nil)
	stream.On("Delete", "run-name").Return(nil)
	keyValueStore.On("Delete", "url:run-name").Return(nil)
	keyValueStore.On("Delete", "token:run-name").Return(nil)

	app := AppImplementation{
		JobDispatcher: jobDispatcher,
		BlobStore:     blobStore,
		Stream:        stream,
		KeyValueStore: keyValueStore,
	}
	err := app.DeleteRun("run-name")
	assert.Nil(t, err)

	jobDispatcher.AssertExpectations(t)
	blobStore.AssertExpectations(t)
	stream.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
}
