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
	"github.com/openaustralia/yinyo/pkg/jobdispatcher"
	"github.com/openaustralia/yinyo/pkg/keyvaluestore"
	"github.com/openaustralia/yinyo/pkg/protocol"
	"github.com/openaustralia/yinyo/pkg/stream"
)

const dockerImage = "openaustralia/yinyo-scraper:v1"
const runBinary = "/bin/yinyo"

// App is the interface for the operations of the server
type App interface {
	CreateRun(namePrefix string) (protocol.Run, error)
	DeleteRun(runName string) error
	StartRun(runName string, output string, env map[string]string, callbackURL string, maxRunTime int64) error
	GetApp(runName string) (io.Reader, error)
	PutApp(reader io.Reader, objectSize int64, runName string) error
	GetCache(runName string) (io.Reader, error)
	PutCache(reader io.Reader, objectSize int64, runName string) error
	GetOutput(runName string) (io.Reader, error)
	PutOutput(reader io.Reader, objectSize int64, runName string) error
	GetExitData(runName string) (protocol.ExitData, error)
	PutExitData(runName string, exitData protocol.ExitData) error
	GetEvents(runName string, lastID string) EventIterator
	CreateEvent(runName string, event protocol.Event) error
	GetTokenCache(runName string) (string, error)
}

type EventIterator interface {
	More() bool
	Next() (e protocol.Event, err error)
}

// AppImplementation holds the state for the application
type AppImplementation struct {
	BlobStore     blobstore.BlobStore
	JobDispatcher jobdispatcher.Jobs
	Stream        stream.Stream
	KeyValueStore keyvaluestore.KeyValueStore
	HTTP          *http.Client
}

type StartupOptions struct {
	Minio MinioOptions
	Redis RedisOptions
}

type MinioOptions struct {
	Host      string
	Bucket    string
	AccessKey string
	SecretKey string
}

type RedisOptions struct {
	Address  string
	Password string
}

// New initialises the main state of the application
func New(startupOptions StartupOptions) (App, error) {
	storeAccess, err := blobstore.NewMinioClient(
		startupOptions.Minio.Host,
		startupOptions.Minio.Bucket,
		startupOptions.Minio.AccessKey,
		startupOptions.Minio.SecretKey,
	)
	if err != nil {
		return nil, err
	}

	// Connect to redis and initially just check that we can connect
	redisClient := redis.NewClient(&redis.Options{
		Addr:     startupOptions.Redis.Address,
		Password: startupOptions.Redis.Password,
	})
	_, err = redisClient.Ping().Result()
	if err != nil {
		return nil, err
	}

	streamClient := stream.NewRedis(redisClient)

	jobDispatcher, err := jobdispatcher.NewKubernetes()
	if err != nil {
		return nil, err
	}

	keyValueStore := keyvaluestore.NewRedis(redisClient)

	return &AppImplementation{
		BlobStore:     storeAccess,
		JobDispatcher: jobDispatcher,
		Stream:        streamClient,
		KeyValueStore: keyValueStore,
		HTTP:          http.DefaultClient,
	}, nil
}

// CreateRun creates a run
func (app *AppImplementation) CreateRun(namePrefix string) (protocol.Run, error) {
	var createResult protocol.Run
	if namePrefix == "" {
		namePrefix = "run"
	}
	// Generate random token
	runToken := uniuri.NewLen(32)
	runName, err := app.JobDispatcher.CreateJobAndToken(namePrefix, runToken)
	if err != nil {
		return createResult, err
	}

	createResult = protocol.Run{
		Name:  runName,
		Token: runToken,
	}

	// Now cache the token for quicker access
	err = app.setKeyValueData(runName, tokenCacheKey, runToken)
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
	r, err := app.getKeyValueData(runName, exitDataKey)
	if err != nil {
		return exitData, err
	}
	err = json.Unmarshal([]byte(r), &exitData)
	return exitData, err
}

// PutExitData uploads the exit data
func (app *AppImplementation) PutExitData(runName string, exitData protocol.ExitData) error {
	b, err := json.Marshal(exitData)
	if err != nil {
		return err
	}
	return app.setKeyValueData(runName, exitDataKey, string(b))
}

func (app *AppImplementation) MaxRunTime() int64 {
	// TODO: Make this configurable at the server level
	// 24 hours in seconds
	return 86400
}

// StartRun starts the run
func (app *AppImplementation) StartRun(
	runName string, output string, env map[string]string, callbackURL string, maxRunTime int64) error {
	if maxRunTime == 0 {
		maxRunTime = app.MaxRunTime()
	} else if maxRunTime > app.MaxRunTime() {
		return ErrMaxRunTimeTooLarge
	}

	// First check that the app exists
	_, err := app.GetApp(runName)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrAppNotAvailable
		}
		return err
	}

	err = app.setKeyValueData(runName, callbackKey, callbackURL)
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
	return app.JobDispatcher.StartJob(runName, dockerImage, command, maxRunTime)
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
func (app *AppImplementation) GetEvents(runName string, lastID string) EventIterator {
	return &Events{app: app, runName: runName, lastID: lastID, more: true}
}

// More checks whether there are more events available. If true you can then call Next()
func (events *Events) More() bool {
	return events.more
}

// Next returns the next event
func (events *Events) Next() (e protocol.Event, err error) {
	e, err = events.app.Stream.Get(events.runName, events.lastID)
	if err != nil {
		return
	}

	// Add the id to the event
	events.lastID = e.ID

	// Check if this is the last event
	_, ok := e.Data.(protocol.LastData)
	events.more = !ok
	return
}

// CreateEvent add an event to the stream
func (app *AppImplementation) CreateEvent(runName string, event protocol.Event) error {
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
	err = app.deleteKeyValueData(runName, exitDataKey)
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
	err = app.deleteKeyValueData(runName, callbackKey)
	if err != nil {
		return err
	}
	return app.deleteKeyValueData(runName, tokenCacheKey)
}

// GetTokenCache gets the cached runToken. Returns ErrNotFound if run name doesn't exist
func (app *AppImplementation) GetTokenCache(runName string) (string, error) {
	return app.getKeyValueData(runName, tokenCacheKey)
}

func (app *AppImplementation) postCallbackEvent(runName string, event protocol.Event) error {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	err := enc.Encode(event)
	if err != nil {
		return err
	}

	callbackURL, err := app.getKeyValueData(runName, callbackKey)
	if err != nil {
		return err
	}

	// Only do the callback if there's a sensible URL
	if callbackURL != "" {
		resp, err := app.HTTP.Post(callbackURL, "application/json", &b)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return errors.New("callback: " + resp.Status)
		}
	}
	return nil
}
