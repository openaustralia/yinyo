package commands

import (
	"bytes"
	"encoding/csv"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/dchest/uniuri"
	"github.com/go-redis/redis"

	"github.com/openaustralia/yinyo/pkg/blobstore"
	"github.com/openaustralia/yinyo/pkg/event"
	"github.com/openaustralia/yinyo/pkg/jobdispatcher"
	"github.com/openaustralia/yinyo/pkg/keyvaluestore"
	"github.com/openaustralia/yinyo/pkg/stream"
)

const filenameApp = "app.tgz"
const filenameCache = "cache.tgz"
const filenameOutput = "output"
const filenameExitData = "exit-data.json"
const dockerImage = "openaustralia/yinyo-scraper:v1"
const runBinary = "/bin/yinyo"

// App holds the state for the application
type App struct {
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
func New() (*App, error) {
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

	return &App{
		BlobStore:     storeAccess,
		JobDispatcher: jobDispatcher,
		Stream:        streamClient,
		KeyValueStore: keyValueStore,
		HTTP:          defaultHTTP(),
	}, nil
}

// CreateRun creates a run
func (app *App) CreateRun(namePrefix string) (CreateRunResult, error) {
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
func (app *App) GetApp(runName string) (io.Reader, error) {
	return app.getData(runName, filenameApp)
}

// PutApp uploads the tar & gzipped application code
func (app *App) PutApp(reader io.Reader, objectSize int64, runName string) error {
	return app.putData(reader, objectSize, runName, filenameApp)
}

// GetCache downloads the tar & gzipped build cache
func (app *App) GetCache(runName string) (io.Reader, error) {
	return app.getData(runName, filenameCache)
}

// PutCache uploads the tar & gzipped build cache
func (app *App) PutCache(reader io.Reader, objectSize int64, runName string) error {
	return app.putData(reader, objectSize, runName, filenameCache)
}

// GetOutput downloads the scraper output
func (app *App) GetOutput(runName string) (io.Reader, error) {
	return app.getData(runName, filenameOutput)
}

// PutOutput uploads the scraper output
func (app *App) PutOutput(reader io.Reader, objectSize int64, runName string) error {
	return app.putData(reader, objectSize, runName, filenameOutput)
}

// GetExitData downloads the json exit data
func (app *App) GetExitData(runName string) (io.Reader, error) {
	return app.getData(runName, filenameExitData)
}

// PutExitData uploads the (already serialised) json exit data
func (app *App) PutExitData(reader io.Reader, objectSize int64, runName string) error {
	return app.putData(reader, objectSize, runName, filenameExitData)
}

// StartRun starts the run
func (app *App) StartRun(
	runName string, output string, env map[string]string, callbackURL string,
) error {
	// First check that the app exists
	_, err := app.GetApp(runName)
	if err != nil {
		// TODO: Make this a client error
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
	app     *App
	runName string
	lastID  string
	more    bool
}

// GetEvents returns an iterator to get at all the events.
// Use "0" for lastId to start at the beginning of the stream. Otherwise use the id of the last
// seen event to restart the stream from that point. Don't try to restart the stream from the
// last event, otherwise More() will just wait around forever.
func (app *App) GetEvents(runName string, lastID string) Events {
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
func (app *App) CreateEvent(runName string, event event.Event) error {
	// TODO: Use something like runName-events instead for the stream name
	event, err := app.Stream.Add(runName, event)
	if err != nil {
		return err
	}
	return app.postCallbackEvent(runName, event)
}

// DeleteRun deletes the run. Should be the last thing called
func (app *App) DeleteRun(runName string) error {
	err := app.JobDispatcher.DeleteJobAndToken(runName)
	if err != nil {
		return err
	}

	err = app.deleteData(runName, filenameApp)
	if err != nil {
		return err
	}
	err = app.deleteData(runName, filenameOutput)
	if err != nil {
		return err
	}
	err = app.deleteData(runName, filenameExitData)
	if err != nil {
		return err
	}
	err = app.deleteData(runName, filenameCache)
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

func storagePath(runName string, fileName string) string {
	return runName + "/" + fileName
}

func (app *App) getData(runName string, fileName string) (io.Reader, error) {
	return app.BlobStore.Get(storagePath(runName, fileName))
}

func (app *App) putData(reader io.Reader, objectSize int64, runName string, fileName string) error {
	return app.BlobStore.Put(
		storagePath(runName, fileName),
		reader,
		objectSize,
	)
}

func (app *App) deleteData(runName string, fileName string) error {
	return app.BlobStore.Delete(storagePath(runName, fileName))
}
