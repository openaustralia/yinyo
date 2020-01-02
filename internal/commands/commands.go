package commands

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/dchest/uniuri"
	"github.com/go-redis/redis"

	"github.com/openaustralia/yinyo/pkg/archive"
	"github.com/openaustralia/yinyo/pkg/blobstore"
	"github.com/openaustralia/yinyo/pkg/event"
	"github.com/openaustralia/yinyo/pkg/jobdispatcher"
	"github.com/openaustralia/yinyo/pkg/keyvaluestore"
	"github.com/openaustralia/yinyo/pkg/protocol"
	"github.com/openaustralia/yinyo/pkg/stream"
)

const filenameApp = "app.tgz"
const filenameCache = "cache.tgz"
const filenameOutput = "output"
const filenameExitData = "exit-data.json"
const dockerImage = "openaustralia/yinyo-scraper:v1"
const runBinary = "/bin/yinyo"

// App is the interface for the operations of the server
type App interface {
	CreateRun(namePrefix string) (CreateRunResult, error)
	DeleteRun(runName string) error
	StartRun(runName string, output string, env map[string]string, callbackURL string) error
	GetApp(runName string) (io.Reader, error)
	PutApp(reader io.Reader, objectSize int64, runName string) error
	GetCache(runName string) (io.Reader, error)
	PutCache(reader io.Reader, objectSize int64, runName string) error
	GetOutput(runName string) (io.Reader, error)
	PutOutput(reader io.Reader, objectSize int64, runName string) error
	GetExitData(runName string) (protocol.ExitData, error)
	PutExitData(runName string, exitData protocol.ExitData) error
	GetEvents(runName string, lastID string) Events
	CreateEvent(runName string, event event.Event) error
	GetTokenCache(runName string) (string, error)
}

// AppImplementation holds the state for the application
type AppImplementation struct {
	BlobStore     blobstore.Client
	JobDispatcher jobdispatcher.Client
	Stream        stream.Client
	KeyValueStore keyvaluestore.Client
	HTTP          *http.Client
}

// CreateRunResult is the output of CreateRun
type CreateRunResult struct {
	RunName  string `json:"name"`
	RunToken string `json:"token"`
}

type logMessage struct {
	// TODO: Make the stream, stage and type an enum
	Log, Stream, Stage, Type string
}

func defaultStore() (blobstore.Client, error) {
	return blobstore.NewMinioClient(
		os.Getenv("STORE_HOST"),
		os.Getenv("STORE_BUCKET"),
		os.Getenv("STORE_ACCESS_KEY"),
		os.Getenv("STORE_SECRET_KEY"),
	)
}

func defaultRedis() (*redis.Client, error) {
	// Connect to redis and initially just check that we can connect
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: os.Getenv("REDIS_PASSWORD"),
	})
	_, err := redisClient.Ping().Result()
	if err != nil {
		return nil, err
	}
	return redisClient, nil
}

func defaultJobDispatcher() (jobdispatcher.Client, error) {
	return jobdispatcher.NewKubernetes()
}

func defaultHTTP() *http.Client {
	return http.DefaultClient
}

// New initialises the main state of the application
func New() (App, error) {
	storeAccess, err := defaultStore()
	if err != nil {
		return nil, err
	}

	redisClient, err := defaultRedis()
	if err != nil {
		return nil, err
	}
	streamClient := stream.NewRedis(redisClient)

	jobDispatcher, err := defaultJobDispatcher()
	if err != nil {
		return nil, err
	}

	keyValueStore := keyvaluestore.NewRedis(redisClient)

	return &AppImplementation{
		BlobStore:     storeAccess,
		JobDispatcher: jobDispatcher,
		Stream:        streamClient,
		KeyValueStore: keyValueStore,
		HTTP:          defaultHTTP(),
	}, nil
}

// CreateRun creates a run
func (app *AppImplementation) CreateRun(namePrefix string) (CreateRunResult, error) {
	if namePrefix == "" {
		namePrefix = "run"
	}
	// Generate random token
	runToken := uniuri.NewLen(32)
	runName, err := app.JobDispatcher.CreateJobAndToken(namePrefix, runToken)

	// Now cache the token for quicker access
	app.setTokenCache(runName, runToken)

	createResult := CreateRunResult{
		RunName:  runName,
		RunToken: runToken,
	}
	return createResult, err
}

// GetApp downloads the tar & gzipped application code
func (app *AppImplementation) GetApp(runName string) (io.Reader, error) {
	return app.getBlobStoreData(runName, filenameApp)
}

// Simultaneously check that the archive is valid and save to a temporary file
// If this errors the temp file will not be created
// Responsibility of the caller to delete the temporary file
func (app *AppImplementation) validateArchiveToTempFile(reader io.Reader) (*os.File, error) {
	tmpfile, err := ioutil.TempFile("", filenameApp)
	if err != nil {
		return tmpfile, err
	}

	r := io.TeeReader(reader, tmpfile)
	err = archive.Validate(r)
	if err != nil {
		os.Remove(tmpfile.Name())
		return tmpfile, fmt.Errorf("%w: %v", ErrArchiveFormat, err)
	}

	// Go back to the beginning of the temporary file
	_, err = tmpfile.Seek(0, io.SeekStart)
	if err != nil {
		os.Remove(tmpfile.Name())
		return tmpfile, err
	}
	return tmpfile, nil
}

// PutApp uploads the tar & gzipped application code
func (app *AppImplementation) PutApp(reader io.Reader, objectSize int64, runName string) error {
	tmpfile, err := app.validateArchiveToTempFile(reader)
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())

	// Now upload the contents of the temporary file
	return app.putBlobStoreData(tmpfile, objectSize, runName, filenameApp)
}

// GetCache downloads the tar & gzipped build cache
func (app *AppImplementation) GetCache(runName string) (io.Reader, error) {
	return app.getBlobStoreData(runName, filenameCache)
}

// PutCache uploads the tar & gzipped build cache
func (app *AppImplementation) PutCache(reader io.Reader, objectSize int64, runName string) error {
	tmpfile, err := app.validateArchiveToTempFile(reader)
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())

	return app.putBlobStoreData(tmpfile, objectSize, runName, filenameCache)
}

// GetOutput downloads the scraper output
func (app *AppImplementation) GetOutput(runName string) (io.Reader, error) {
	return app.getBlobStoreData(runName, filenameOutput)
}

// PutOutput uploads the scraper output
func (app *AppImplementation) PutOutput(reader io.Reader, objectSize int64, runName string) error {
	return app.putBlobStoreData(reader, objectSize, runName, filenameOutput)
}

// GetExitData downloads the exit data
func (app *AppImplementation) GetExitData(runName string) (protocol.ExitData, error) {
	var exitData protocol.ExitData
	reader, err := app.getBlobStoreData(runName, filenameExitData)
	if err != nil {
		return exitData, err
	}

	dec := json.NewDecoder(reader)
	err = dec.Decode(&exitData)
	return exitData, err
}

// PutExitData uploads the exit data
// TODO: Store this in redis rather than on the blobstore
func (app *AppImplementation) PutExitData(runName string, exitData protocol.ExitData) error {
	b, err := json.Marshal(exitData)
	if err != nil {
		return err
	}
	return app.putBlobStoreData(strings.NewReader(string(b)), int64(len(b)), runName, filenameExitData)
}

// StartRun starts the run
func (app *AppImplementation) StartRun(
	runName string, output string, env map[string]string, callbackURL string,
) error {
	// First check that the app exists
	_, err := app.GetApp(runName)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrAppNotAvailable
		}
		return err
	}

	err = app.setCallbackURL(runName, callbackURL)
	if err != nil {
		return err
	}
	runToken, err := app.JobDispatcher.GetToken(runName)
	if err != nil {
		return err
	}

	// Convert environment variable values to a single string that can be passed
	// as a flag to wrapper
	records := make([]string, 0, len(env)>>1)
	for k, v := range env {
		records = append(records, k+"="+v)
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write(records); err != nil {
		return err
	}
	w.Flush()

	envString := strings.TrimSpace(buf.String())
	command := []string{
		runBinary,
		"wrapper",
		runName,
		runToken,
		"--output", output,
	}
	if envString != "" {
		command = append(command, "--env", envString)
	}
	return app.JobDispatcher.StartJob(runName, dockerImage, command)
}

// Events is an iterator to retrieve events from a stream
type Events struct {
	app     *AppImplementation
	runName string
	lastID  string
	more    bool
}

// GetEvents returns an iterator to get at all the events.
// Use "0" for lastId to start at the beginning of the stream. Otherwise use the id of the last
// seen event to restart the stream from that point. Don't try to restart the stream from the
// last event, otherwise More() will just wait around forever.
func (app *AppImplementation) GetEvents(runName string, lastID string) Events {
	return Events{app: app, runName: runName, lastID: lastID, more: true}
}

// More checks whether there are more events available. If true you can then call Next()
func (events *Events) More() bool {
	return events.more
}

// Next returns the next event
func (events *Events) Next() (e event.Event, err error) {
	e, err = events.app.Stream.Get(events.runName, events.lastID)
	if err != nil {
		return
	}

	// Add the id to the event
	events.lastID = e.ID

	// Check if this is the last event
	_, ok := e.Data.(event.LastData)
	events.more = !ok
	return
}

// CreateEvent add an event to the stream
func (app *AppImplementation) CreateEvent(runName string, event event.Event) error {
	// TODO: Use something like runName-events instead for the stream name
	event, err := app.Stream.Add(runName, event)
	if err != nil {
		return err
	}
	return app.postCallbackEvent(runName, event)
}

// DeleteRun deletes the run. Should be the last thing called
func (app *AppImplementation) DeleteRun(runName string) error {
	err := app.JobDispatcher.DeleteJobAndToken(runName)
	if err != nil {
		return err
	}

	err = app.deleteBlobStoreData(runName, filenameApp)
	if err != nil {
		return err
	}
	err = app.deleteBlobStoreData(runName, filenameOutput)
	if err != nil {
		return err
	}
	err = app.deleteBlobStoreData(runName, filenameExitData)
	if err != nil {
		return err
	}
	err = app.deleteBlobStoreData(runName, filenameCache)
	if err != nil {
		return err
	}
	err = app.Stream.Delete(runName)
	if err != nil {
		return err
	}
	err = app.deleteCallbackURL(runName)
	if err != nil {
		return err
	}
	return app.deleteTokenCache(runName)
}

func blobStoreStoragePath(runName string, fileName string) string {
	return runName + "/" + fileName
}

func (app *AppImplementation) getBlobStoreData(runName string, fileName string) (io.Reader, error) {
	p := blobStoreStoragePath(runName, fileName)
	r, err := app.BlobStore.Get(p)
	if err != nil && app.BlobStore.IsNotExist(err) {
		return r, fmt.Errorf("blobstore %v: %w", p, ErrNotFound)
	}
	return r, err
}

func (app *AppImplementation) putBlobStoreData(reader io.Reader, objectSize int64, runName string, fileName string) error {
	return app.BlobStore.Put(
		blobStoreStoragePath(runName, fileName),
		reader,
		objectSize,
	)
}

func (app *AppImplementation) deleteBlobStoreData(runName string, fileName string) error {
	return app.BlobStore.Delete(blobStoreStoragePath(runName, fileName))
}
