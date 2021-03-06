package commands

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	blobstoremocks "github.com/openaustralia/yinyo/mocks/pkg/blobstore"
	jobdispatchermocks "github.com/openaustralia/yinyo/mocks/pkg/jobdispatcher"
	keyvaluestoremocks "github.com/openaustralia/yinyo/mocks/pkg/keyvaluestore"
	streammocks "github.com/openaustralia/yinyo/mocks/pkg/stream"
	"github.com/openaustralia/yinyo/pkg/integrationclient"
	"github.com/openaustralia/yinyo/pkg/keyvaluestore"
	"github.com/openaustralia/yinyo/pkg/protocol"
)

func TestStoragePath(t *testing.T) {
	assert.Equal(t, blobStoreStoragePath("abc", "app.tgz"), "abc/app.tgz")
	assert.Equal(t, blobStoreStoragePath("def", "output"), "def/output")
}

func TestStartRun(t *testing.T) {
	job := new(jobdispatchermocks.Jobs)
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)
	blobStore := new(blobstoremocks.BlobStore)

	// Expect that the job will get dispatched
	job.On(
		"Create",
		"run-name",
		"image",
		[]string{"/bin/wrapper", "run-name", "--output", "output.txt", "--server", "http://localhost:8080", "--env", "FOO=bar"},
		int64(86400),
		int64(512*1024*1024),
	).Return(nil)
	// Expect that we save the callback url in the key value store
	keyValueStore.On("Set", "run-name/url", `"http://foo.com"`).Return(nil)
	// Expect that we save away the amount of memory allocated to the run
	keyValueStore.On("Set", "run-name/memory", "536870912").Return(nil)
	// Expect that we try to get the code just to see if it exists
	blobStore.On("Get", "run-name/app.tgz").Return(nil, nil)

	app := AppImplementation{integrationClient: &integrationclient.Client{}, JobDispatcher: job, KeyValueStore: keyValueStore, BlobStore: blobStore, ServerURL: "http://localhost:8080"}
	err := app.StartRun(
		"run-name",
		"image",
		protocol.StartRunOptions{
			Output:     "output.txt",
			Env:        []protocol.EnvVariable{{Name: "FOO", Value: "bar"}},
			Callback:   protocol.Callback{URL: "http://foo.com"},
			MaxRunTime: 86400,
			Memory:     512 * 1024 * 1024,
		},
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
	stream := new(streammocks.Stream)
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)

	time := time.Now()
	stream.On("Add", "run-name", protocol.NewStartEvent("", "abc", time, "build")).Return(protocol.NewStartEvent("123", "abc", time, "build"), nil)
	keyValueStore.On("Get", "run-name/url").Return(`"http://foo.com/bar"`, nil)

	// Mock out the http RoundTripper so that no actual http request is made
	httpClient := http.DefaultClient
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

	app := AppImplementation{integrationClient: &integrationclient.Client{}, Stream: stream, KeyValueStore: keyValueStore, HTTP: httpClient}
	err := app.CreateEvent("run-name", protocol.NewStartEvent("", "abc", time, "build"))
	assert.Nil(t, err)

	stream.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
	roundTripper.AssertExpectations(t)
}

func TestCreateFinishEvent(t *testing.T) {
	time := time.Now()
	stream := new(streammocks.Stream)
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)
	app := AppImplementation{integrationClient: &integrationclient.Client{}, Stream: stream, KeyValueStore: keyValueStore}

	exitData := protocol.ExitDataStage{ExitCode: 12, Usage: protocol.StageUsage{MaxRSS: 100, NetworkIn: 200, NetworkOut: 300}}
	event := protocol.NewFinishEvent("", "abc", time, "build", exitData)
	eventWithID := protocol.NewFinishEvent("123", "abc", time, "build", exitData)

	stream.On("Add", "run-name", event).Return(eventWithID, nil)
	keyValueStore.On("Get", "run-name/url").Return("", nil)
	keyValueStore.On("Set", "run-name/exit_data/build", `{"exit_code":12,"usage":{"max_rss":100,"network_in":200,"network_out":300}}`).Return(nil)

	app.CreateEvent("run-name", event)

	stream.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
}

func TestCreateFirstEvent(t *testing.T) {
	stream := new(streammocks.Stream)
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)
	app := AppImplementation{Stream: stream, KeyValueStore: keyValueStore}

	time := time.Date(2020, 3, 11, 15, 24, 30, 0, time.UTC)
	event := protocol.NewFirstEvent("", "abc", time)
	eventWithID := protocol.NewFirstEvent("123", "abc", time)

	stream.On("Add", "run-name", event).Return(eventWithID, nil)
	keyValueStore.On("Set", "run-name/first_time", `"2020-03-11T15:24:30Z"`).Return(nil)
	keyValueStore.On("Get", "run-name/url").Return("", nil)

	app.CreateEvent("run-name", event)

	stream.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
}

func TestCreateLastEvent(t *testing.T) {
	stream := new(streammocks.Stream)
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)
	app := AppImplementation{integrationClient: &integrationclient.Client{}, Stream: stream, KeyValueStore: keyValueStore}

	time := time.Now()
	event := protocol.NewLastEvent("", "abc", time)
	eventWithID := protocol.NewLastEvent("123", "abc", time)

	stream.On("Add", "run-name", event).Return(eventWithID, nil)
	keyValueStore.On("Get", "run-name/url").Return("", nil)
	keyValueStore.On("Get", "run-name/first_time").Return(`"2020-03-11T15:24:30Z"`, nil)
	keyValueStore.On("Set", "run-name/exit_data/finished", `true`).Return(nil)
	keyValueStore.On("Get", "run-name/memory").Return("1073741824", nil)

	app.CreateEvent("run-name", event)

	stream.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
}

func TestCreateEventNoCallbackURL(t *testing.T) {
	stream := new(streammocks.Stream)
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)

	time := time.Now()
	stream.On("Add", "run-name", protocol.NewStartEvent("", "abc", time, "build")).Return(protocol.NewStartEvent("123", "abc", time, "build"), nil)
	keyValueStore.On("Get", "run-name/url").Return(`""`, nil)

	// Mock out the http RoundTripper so that no actual http request is made
	httpClient := http.DefaultClient
	roundTripper := new(MockRoundTripper)
	httpClient.Transport = roundTripper

	app := AppImplementation{integrationClient: &integrationclient.Client{}, Stream: stream, KeyValueStore: keyValueStore, HTTP: httpClient}
	err := app.CreateEvent("run-name", protocol.NewStartEvent("", "abc", time, "build"))
	assert.Nil(t, err)

	stream.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
	roundTripper.AssertNotCalled(t, "RoundTrip")
}

func TestCreateEventErrorOneTimeDuringCallback(t *testing.T) {
	stream := new(streammocks.Stream)
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)

	time := time.Now()
	stream.On("Add", "run-name", protocol.NewStartEvent("", "abc", time, "build")).Return(protocol.NewStartEvent("123", "abc", time, "build"), nil)
	keyValueStore.On("Get", "run-name/url").Return(`"http://foo.com/bar"`, nil)

	// Mock out the http RoundTripper so that no actual http request is made
	httpClient := http.DefaultClient
	roundTripper := new(MockRoundTripper)
	// Simulating the remote host failing 1 time and then succeeding
	roundTripper.On("RoundTrip", mock.MatchedBy(func(r *http.Request) bool {
		return r.URL.String() == "http://foo.com/bar"
	})).Return(
		&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(strings.NewReader("")),
		},
		nil,
	).Once()
	roundTripper.On("RoundTrip", mock.MatchedBy(func(r *http.Request) bool {
		return r.URL.String() == "http://foo.com/bar"
	})).Return(
		&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader("")),
		},
		nil,
	).Once()
	httpClient.Transport = roundTripper

	app := AppImplementation{integrationClient: &integrationclient.Client{}, Stream: stream, KeyValueStore: keyValueStore, HTTP: httpClient}
	err := app.CreateEvent("run-name", protocol.NewStartEvent("", "abc", time, "build"))
	assert.Nil(t, err)

	stream.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
	roundTripper.AssertExpectations(t)
}
func TestCreateEventErrorFiveTimesDuringCallback(t *testing.T) {
	stream := new(streammocks.Stream)
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)

	time := time.Now()
	stream.On("Add", "run-name", protocol.NewStartEvent("", "abc", time, "build")).Return(protocol.NewStartEvent("123", "abc", time, "build"), nil)
	keyValueStore.On("Get", "run-name/url").Return(`"http://foo.com/bar"`, nil)

	// Mock out the http RoundTripper so that no actual http request is made
	httpClient := http.DefaultClient
	roundTripper := new(MockRoundTripper)
	// Simulating the remote host failing 5 times in a row
	roundTripper.On("RoundTrip", mock.MatchedBy(func(r *http.Request) bool {
		return r.URL.String() == "http://foo.com/bar"
	})).Return(
		&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(strings.NewReader("")),
		},
		nil,
	).Times(5)
	httpClient.Transport = roundTripper

	app := AppImplementation{Stream: stream, KeyValueStore: keyValueStore, HTTP: httpClient}
	err := app.CreateEvent("run-name", protocol.NewStartEvent("", "abc", time, "build"))
	assert.EqualError(t, err, "POST http://foo.com/bar giving up after 5 attempts")

	stream.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
	roundTripper.AssertExpectations(t)
}

func TestGetEvents(t *testing.T) {
	stream := new(streammocks.Stream)

	time := time.Now()
	stream.On("Get", "run-name", "0").Return(protocol.NewStartEvent("123", "abc", time, "build"), nil)
	stream.On("Get", "run-name", "123").Return(protocol.NewLastEvent("456", "abc", time), nil)

	app := AppImplementation{Stream: stream}

	events := app.GetEvents("run-name", "0")

	// We're expecting two events in the stream. Let's hardcode what would normally be in a loop
	assert.True(t, events.More())
	e, err := events.Next()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, protocol.NewStartEvent("123", "abc", time, "build"), e)
	assert.True(t, events.More())
	e, err = events.Next()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, protocol.NewLastEvent("456", "abc", time), e)
	assert.False(t, events.More())

	stream.AssertExpectations(t)
}

func TestDeleteRun(t *testing.T) {
	jobDispatcher := new(jobdispatchermocks.Jobs)
	blobStore := new(blobstoremocks.BlobStore)
	stream := new(streammocks.Stream)
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)

	jobDispatcher.On("Delete", "run-name").Return(nil)
	blobStore.On("Delete", "run-name/app.tgz").Return(nil)
	blobStore.On("Delete", "run-name/output").Return(nil)
	blobStore.On("Delete", "run-name/cache.tgz").Return(nil)
	stream.On("Delete", "run-name").Return(nil)
	keyValueStore.On("Delete", "run-name/url").Return(nil)
	keyValueStore.On("Delete", "run-name/created").Return(nil)
	keyValueStore.On("Delete", "run-name/first_time").Return(nil)
	keyValueStore.On("Delete", "run-name/memory").Return(nil)
	keyValueStore.On("Delete", "run-name/exit_data/build").Return(nil)
	keyValueStore.On("Delete", "run-name/exit_data/execute").Return(nil)
	keyValueStore.On("Delete", "run-name/exit_data/finished").Return(nil)

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

func TestIsRunCreatedNotFound(t *testing.T) {
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)
	keyValueStore.On("Get", "does-not-exit/created").Return("", keyvaluestore.ErrKeyNotExist)

	app := AppImplementation{KeyValueStore: keyValueStore}
	// This run name should not exist
	created, err := app.IsRunCreated("does-not-exit")
	assert.Nil(t, err)
	assert.False(t, created)

	keyValueStore.AssertExpectations(t)
}

func TestPutApp(t *testing.T) {
	blobStore := new(blobstoremocks.BlobStore)
	app := AppImplementation{BlobStore: blobStore}

	blobStore.On("Put", "run-name/app.tgz", mock.Anything, mock.Anything).Return(nil)

	// Open a file which has the simplest possible archive which is empty but valid
	file, _ := os.Open("testdata/empty.tgz")
	defer file.Close()
	stat, _ := file.Stat()

	err := app.PutApp("run-name", file, stat.Size())
	if err != nil {
		t.Fatal(err)
	}

	blobStore.AssertExpectations(t)
}

func TestGetCache(t *testing.T) {
	blobStore := new(blobstoremocks.BlobStore)
	app := AppImplementation{BlobStore: blobStore}

	file, _ := os.Open("testdata/empty.tgz")
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
	blobStore := new(blobstoremocks.BlobStore)
	app := AppImplementation{BlobStore: blobStore}

	file, _ := os.Open("testdata/empty.tgz")
	stat, _ := file.Stat()

	blobStore.On("Put", "run-name/cache.tgz", mock.Anything, stat.Size()).Return(nil)

	err := app.PutCache("run-name", file, stat.Size())
	if err != nil {
		t.Fatal(err)
	}

	blobStore.AssertExpectations(t)
}

func TestGetOutput(t *testing.T) {
	blobStore := new(blobstoremocks.BlobStore)
	app := AppImplementation{BlobStore: blobStore}

	blobStore.On("Get", "run-name/output").Return(strings.NewReader("output"), nil)

	r, err := app.GetOutput("run-name")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := ioutil.ReadAll(r)

	assert.Equal(t, "output", string(b))
	blobStore.AssertExpectations(t)
}

func TestPutOutput(t *testing.T) {
	blobStore := new(blobstoremocks.BlobStore)
	app := AppImplementation{BlobStore: blobStore}

	reader := strings.NewReader("output")
	blobStore.On("Put", "run-name/output", reader, int64(6)).Return(nil)

	err := app.PutOutput("run-name", reader, 6)
	if err != nil {
		t.Fatal(err)
	}
	blobStore.AssertExpectations(t)
}

func TestGetExitData(t *testing.T) {
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)
	app := AppImplementation{KeyValueStore: keyValueStore}

	keyValueStore.On("Get", "run-name/exit_data/build").Return(`{"exit_code":0,"usage":{"max_rss":1,"network_in":0,"network_out":0}}`, nil)
	keyValueStore.On("Get", "run-name/exit_data/execute").Return(`{"exit_code":0,"usage":{"max_rss":2,"network_in":0,"network_out":0}}`, nil)
	keyValueStore.On("Get", "run-name/exit_data/finished").Return("true", nil)
	e, err := app.GetExitData("run-name")
	if err != nil {
		t.Fatal(err)
	}
	expectedExitData := protocol.ExitData{
		Build:    &protocol.ExitDataStage{ExitCode: 0, Usage: protocol.StageUsage{MaxRSS: 1}},
		Execute:  &protocol.ExitDataStage{ExitCode: 0, Usage: protocol.StageUsage{MaxRSS: 2}},
		Finished: true,
	}

	assert.Equal(t, expectedExitData, e)
	keyValueStore.AssertExpectations(t)
}

func TestGetExitDataBuildErrored(t *testing.T) {
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)
	app := AppImplementation{KeyValueStore: keyValueStore}

	keyValueStore.On("Get", "run-name/exit_data/build").Return(`{"exit_code":15,"usage":{"max_rss":0,"network_in":0,"network_out":0}}`, nil)
	keyValueStore.On("Get", "run-name/exit_data/execute").Return("", keyvaluestore.ErrKeyNotExist)
	keyValueStore.On("Get", "run-name/exit_data/finished").Return("true", nil)

	e, err := app.GetExitData("run-name")
	if err != nil {
		t.Fatal(err)
	}
	expectedExitData := protocol.ExitData{
		Build:    &protocol.ExitDataStage{ExitCode: 15},
		Finished: true,
	}

	assert.Equal(t, expectedExitData, e)
	keyValueStore.AssertExpectations(t)
}

func TestGetExitDataRunNotStarted(t *testing.T) {
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)
	app := AppImplementation{KeyValueStore: keyValueStore}

	keyValueStore.On("Get", "run-name/exit_data/build").Return("", keyvaluestore.ErrKeyNotExist)
	keyValueStore.On("Get", "run-name/exit_data/execute").Return("", keyvaluestore.ErrKeyNotExist)
	keyValueStore.On("Get", "run-name/exit_data/finished").Return("", keyvaluestore.ErrKeyNotExist)

	e, err := app.GetExitData("run-name")
	if err != nil {
		t.Fatal(err)
	}
	expectedExitData := protocol.ExitData{
		Finished: false,
	}

	assert.Equal(t, expectedExitData, e)
	keyValueStore.AssertExpectations(t)
}

func TestCreateRun(t *testing.T) {
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)
	app := AppImplementation{integrationClient: &integrationclient.Client{}, KeyValueStore: keyValueStore}

	keyValueStore.On("Set", mock.Anything, mock.Anything).Return(nil)

	run, err := app.CreateRun(protocol.CreateRunOptions{})
	if err != nil {
		t.Fatal(err)
	}
	// run.ID should be a uuid. Check that it is
	_, err = uuid.FromString(run.ID)
	assert.Nil(t, err)
	keyValueStore.AssertExpectations(t)
}

func TestCreateRunWithAuthenticationAllowed(t *testing.T) {
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)

	// Mock out the http RoundTripper so that no actual http request is made
	httpClient := http.DefaultClient
	roundTripper := new(MockRoundTripper)
	httpClient.Transport = roundTripper

	integrationClient := integrationclient.New(httpClient, "http://foo.com/authenticate", "", "")
	app := AppImplementation{integrationClient: integrationClient, HTTP: httpClient, KeyValueStore: keyValueStore}

	// Mock a response where the authentication allowed things to go ahead
	roundTripper.On("RoundTrip", mock.MatchedBy(func(r *http.Request) bool {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			t.Fatal(err)
		}
		return r.URL.Host == "foo.com" && values.Get("api_key") == "foobar" && values.Get("run_id") != ""
	})).Return(
		&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(`{"allowed":true}`)),
		},
		nil,
	)

	keyValueStore.On("Set", mock.Anything, "true").Return(nil)

	_, err := app.CreateRun(protocol.CreateRunOptions{APIKey: "foobar"})
	if err != nil {
		t.Fatal(err)
	}

	roundTripper.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
}

func TestCreateRunWithAuthenticationNotAllowed(t *testing.T) {
	keyValueStore := new(keyvaluestoremocks.KeyValueStore)

	// Mock out the http RoundTripper so that no actual http request is made
	httpClient := http.DefaultClient
	roundTripper := new(MockRoundTripper)
	httpClient.Transport = roundTripper

	integrationClient := integrationclient.New(httpClient, "http://foo.com/authenticate", "", "")
	app := AppImplementation{integrationClient: integrationClient, HTTP: httpClient, KeyValueStore: keyValueStore}

	// Mock a response where the authentication allowed things to go ahead
	roundTripper.On("RoundTrip", mock.MatchedBy(func(r *http.Request) bool {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			t.Fatal(err)
		}
		return r.URL.Host == "foo.com" && values.Get("api_key") == "foobar" && values.Get("run_id") != ""
	})).Return(
		&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(`{"allowed":false,"message":"an error"}`)),
		},
		nil,
	)

	_, err := app.CreateRun(protocol.CreateRunOptions{APIKey: "foobar"})
	assert.True(t, errors.Is(err, integrationclient.ErrNotAllowed))
	assert.Equal(t, "Not allowed: an error", err.Error())

	roundTripper.AssertExpectations(t)
	keyValueStore.AssertExpectations(t)
}

// If we haven't uploaded an app error when starting a run
func TestStartNoApp(t *testing.T) {
	blobstoreClient := new(blobstoremocks.BlobStore)
	app := AppImplementation{BlobStore: blobstoreClient}

	blobstoreClient.On("Get", "foo/app.tgz").Return(nil, errors.New("Doesn't exist"))
	blobstoreClient.On("IsNotExist", errors.New("Doesn't exist")).Return(true)

	err := app.StartRun("foo", "image", protocol.StartRunOptions{})
	assert.True(t, errors.Is(err, ErrAppNotAvailable))

	blobstoreClient.AssertExpectations(t)
}
