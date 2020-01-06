package commands

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	blobstoremocks "github.com/openaustralia/yinyo/mocks/pkg/blobstore"
	jobdispatchermocks "github.com/openaustralia/yinyo/mocks/pkg/jobdispatcher"
	keyvaluestoremocks "github.com/openaustralia/yinyo/mocks/pkg/keyvaluestore"
	streammocks "github.com/openaustralia/yinyo/mocks/pkg/stream"
	"github.com/openaustralia/yinyo/pkg/event"
	"github.com/openaustralia/yinyo/pkg/keyvaluestore"
)

func TestStoragePath(t *testing.T) {
	assert.Equal(t, blobStoreStoragePath("abc", "app.tgz"), "abc/app.tgz")
	assert.Equal(t, blobStoreStoragePath("def", "output"), "def/output")
}

func TestStartRun(t *testing.T) {
	job := new(jobdispatchermocks.Client)
	keyValueStore := new(keyvaluestoremocks.Client)
	blobStore := new(blobstoremocks.Client)

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
	keyValueStore.On("Set", "run-name/url", "http://foo.com").Return(nil)
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
	stream := new(streammocks.Client)
	keyValueStore := new(keyvaluestoremocks.Client)

	time := time.Now()
	stream.On("Add", "run-name", event.NewStartEvent("", time, "build")).Return(event.NewStartEvent("123", time, "build"), nil)
	keyValueStore.On("Get", "run-name/url").Return("http://foo.com/bar", nil)

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
	stream := new(streammocks.Client)
	keyValueStore := new(keyvaluestoremocks.Client)

	time := time.Now()
	stream.On("Add", "run-name", event.NewStartEvent("", time, "build")).Return(event.NewStartEvent("123", time, "build"), nil)
	keyValueStore.On("Get", "run-name/url").Return("", nil)

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
	stream := new(streammocks.Client)
	keyValueStore := new(keyvaluestoremocks.Client)

	time := time.Now()
	stream.On("Add", "run-name", event.NewStartEvent("", time, "build")).Return(event.NewStartEvent("123", time, "build"), nil)
	keyValueStore.On("Get", "run-name/url").Return("http://foo.com/bar", nil)

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
	stream := new(streammocks.Client)

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
	jobDispatcher := new(jobdispatchermocks.Client)
	blobStore := new(blobstoremocks.Client)
	stream := new(streammocks.Client)
	keyValueStore := new(keyvaluestoremocks.Client)

	jobDispatcher.On("DeleteJobAndToken", "run-name").Return(nil)
	blobStore.On("Delete", "run-name/app.tgz").Return(nil)
	blobStore.On("Delete", "run-name/output").Return(nil)
	blobStore.On("Delete", "run-name/cache.tgz").Return(nil)
	stream.On("Delete", "run-name").Return(nil)
	keyValueStore.On("Delete", "run-name/url").Return(nil)
	keyValueStore.On("Delete", "run-name/token").Return(nil)
	keyValueStore.On("Delete", "run-name/exit_data").Return(nil)

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

func TestTokenCacheNotFound(t *testing.T) {
	keyValueStore := new(keyvaluestoremocks.Client)
	keyValueStore.On("Get", "does-not-exit/token").Return("", keyvaluestore.ErrKeyNotExist)

	app := AppImplementation{KeyValueStore: keyValueStore}
	// This run name should not exist
	_, err := app.GetTokenCache("does-not-exit")
	assert.True(t, errors.Is(err, ErrNotFound))

	keyValueStore.AssertExpectations(t)
}

func TestPutApp(t *testing.T) {
	blobStore := new(blobstoremocks.Client)
	app := AppImplementation{BlobStore: blobStore}

	blobStore.On("Put", "run-name/app.tgz", mock.Anything, mock.Anything).Return(nil)

	// Open a file which has the simplest possible archive which is empty but valid
	file, err := os.Open("testdata/empty.tgz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}

	err = app.PutApp(file, stat.Size(), "run-name")
	if err != nil {
		t.Fatal(err)
	}

	blobStore.AssertExpectations(t)
}

func TestGetCache(t *testing.T) {
	blobStore := new(blobstoremocks.Client)
	app := AppImplementation{BlobStore: blobStore}

	file, err := os.Open("testdata/empty.tgz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	blobStore.On("Get", "run-name/cache.tgz").Return(file, nil)

	r, err := app.GetCache("run-name")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, file, r)

	blobStore.AssertExpectations(t)
}

func TestPutCache(t *testing.T) {
	blobStore := new(blobstoremocks.Client)
	app := AppImplementation{BlobStore: blobStore}

	file, _ := os.Open("testdata/empty.tgz")
	stat, _ := file.Stat()

	blobStore.On("Put", "run-name/cache.tgz", mock.Anything, stat.Size()).Return(nil)

	err := app.PutCache(file, stat.Size(), "run-name")
	if err != nil {
		t.Fatal(err)
	}

	blobStore.AssertExpectations(t)
}
