package commands

import (
	"io"
	"os"

	"github.com/dchest/uniuri"

	"github.com/openaustralia/morph-ng/pkg/jobdispatcher"
	"github.com/openaustralia/morph-ng/pkg/store"
	"github.com/openaustralia/morph-ng/pkg/stream"
)

const filenameApp = "app.tgz"
const filenameCache = "cache.tgz"
const filenameOutput = "output"
const filenameExitData = "exit-data.json"
const dockerImage = "openaustralia/clay-scraper:v1"
const runBinary = "/bin/run.sh"

// App holds the state for the application
type App struct {
	Store  store.Client
	Job    jobdispatcher.Client
	Stream stream.Stream
}

// CreateRunResult is the output of CreateRun
type CreateRunResult struct {
	RunName  string `json:"run_name"`
	RunToken string `json:"run_token"`
}

type logMessage struct {
	// TODO: Make the stream, stage and type an enum
	Log, Stream, Stage, Type string
}

// New initialises the main state of the application
func New() (*App, error) {
	storeAccess, err := store.NewMinioClient(
		// TODO: Get data store url for configmap
		"minio-service:9000",
		// TODO: Make bucket name configurable
		"clay",
		os.Getenv("STORE_ACCESS_KEY"),
		os.Getenv("STORE_SECRET_KEY"),
	)
	if err != nil {
		return nil, err
	}

	streamClient, err := stream.NewRedis(
		"redis:6379",
		os.Getenv("REDIS_PASSWORD"),
	)
	if err != nil {
		return nil, err
	}

	jobDispatcher, err := jobdispatcher.NewKubernetes()
	if err != nil {
		return nil, err
	}

	return &App{Store: storeAccess, Job: jobDispatcher, Stream: streamClient}, nil
}

// CreateRun creates a run
func (app *App) CreateRun(namePrefix string) (CreateRunResult, error) {
	if namePrefix == "" {
		namePrefix = "run"
	}
	// Generate random token
	runToken := uniuri.NewLen(32)
	runName, err := app.Job.CreateJobAndToken(namePrefix, runToken)

	createResult := CreateRunResult{
		RunName:  runName,
		RunToken: runToken,
	}
	return createResult, err
}

// GetApp downloads the tar & gzipped application code
func (app *App) GetApp(runName string, w io.Writer) error {
	return app.getData(runName, filenameApp, w)
}

// PutApp uploads the tar & gzipped application code
func (app *App) PutApp(reader io.Reader, objectSize int64, runName string) error {
	return app.putData(reader, objectSize, runName, filenameApp)
}

// GetCache downloads the tar & gzipped build cache
func (app *App) GetCache(runName string, w io.Writer) error {
	return app.getData(runName, filenameCache, w)
}

// PutCache uploads the tar & gzipped build cache
func (app *App) PutCache(reader io.Reader, objectSize int64, runName string) error {
	return app.putData(reader, objectSize, runName, filenameCache)
}

// GetOutput downloads the scraper output
func (app *App) GetOutput(runName string, w io.Writer) error {
	return app.getData(runName, filenameOutput, w)
}

// PutOutput uploads the scraper output
func (app *App) PutOutput(reader io.Reader, objectSize int64, runName string) error {
	return app.putData(reader, objectSize, runName, filenameOutput)
}

// GetExitData downloads the json exit data
func (app *App) GetExitData(runName string, w io.Writer) error {
	return app.getData(runName, filenameExitData, w)
}

// PutExitData uploads the (already serialised) json exit data
func (app *App) PutExitData(reader io.Reader, objectSize int64, runName string) error {
	return app.putData(reader, objectSize, runName, filenameExitData)
}

// StartRun starts the run
func (app *App) StartRun(runName string, output string, env map[string]string) error {
	command := []string{runBinary, runName, output}
	return app.Job.StartJob(runName, dockerImage, command, env)
}

// GetEvent gets the next event
func (app *App) GetEvent(runName string, id string) (newID string, jsonString string, finished bool, err error) {
	return app.Stream.Get(runName, id)
}

// CreateEvent add an event to the stream
func (app *App) CreateEvent(runName string, eventJSON string) error {
	// TODO: Send the event to the user with an http POST

	// TODO: Use something like runName-events instead for the stream name
	return app.Stream.Add(runName, eventJSON)
}

// DeleteRun deletes the run. Should be the last thing called
func (app *App) DeleteRun(runName string) error {
	err := app.Job.DeleteJobAndToken(runName)
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
	return app.Stream.Delete(runName)
}

func storagePath(runName string, fileName string) string {
	return runName + "/" + fileName
}

func (app *App) getData(runName string, fileName string, writer io.Writer) error {
	reader, err := app.Store.Get(storagePath(runName, fileName))
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, reader)
	return err
}

func (app *App) putData(reader io.Reader, objectSize int64, runName string, fileName string) error {
	return app.Store.Put(
		storagePath(runName, fileName),
		reader,
		objectSize,
	)
}

func (app *App) deleteData(runName string, fileName string) error {
	return app.Store.Delete(storagePath(runName, fileName))
}
