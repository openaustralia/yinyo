package commands

import (
	"bytes"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/openaustralia/yinyo/pkg/archive"
	"github.com/openaustralia/yinyo/pkg/blobstore"
	"github.com/openaustralia/yinyo/pkg/integrationclient"
	"github.com/openaustralia/yinyo/pkg/jobdispatcher"
	"github.com/openaustralia/yinyo/pkg/keyvaluestore"
	"github.com/openaustralia/yinyo/pkg/protocol"
	"github.com/openaustralia/yinyo/pkg/stream"
)

const runBinary = "/bin/wrapper"

// App is the interface for the operations of the server
type App interface {
	CreateRun(options protocol.CreateRunOptions) (protocol.Run, error)
	DeleteRun(runID string) error
	StartRun(runID string, dockerImage string, options protocol.StartRunOptions) error
	GetApp(runID string) (io.Reader, error)
	PutApp(runID string, reader io.Reader, objectSize int64) error
	GetCache(runID string) (io.Reader, error)
	PutCache(runID string, reader io.Reader, objectSize int64) error
	GetOutput(runID string) (io.Reader, error)
	PutOutput(runID string, reader io.Reader, objectSize int64) error
	GetExitData(runID string) (protocol.ExitData, error)
	GetEvents(runID string, lastID string) EventIterator
	CreateEvent(runID string, event protocol.Event) error
	IsRunCreated(runID string) (bool, error)
	ReportAPINetworkUsage(runID string, in uint64, out uint64) error
}

// EventIterator is the interface for getting individual events in a list of events
type EventIterator interface {
	More() bool
	Next() (e protocol.Event, err error)
}

// AppImplementation holds the state for the application
type AppImplementation struct {
	BlobStore         blobstore.BlobStore
	JobDispatcher     jobdispatcher.Jobs
	Stream            stream.Stream
	KeyValueStore     keyvaluestore.KeyValueStore
	HTTP              *http.Client
	integrationClient *integrationclient.Client
}

// StartupOptions are the options available when initialising the application
type StartupOptions struct {
	Minio               MinioOptions
	Redis               RedisOptions
	AuthenticationURL   string
	ResourcesAllowedURL string
	UsageURL            string
}

// MinioOptions are the options for the specific blob storage
type MinioOptions struct {
	Host      string
	Bucket    string
	AccessKey string
	SecretKey string
}

// RedisOptions are the options for the specific key value store
type RedisOptions struct {
	Address  string
	Password string
	TLS      bool
	Database int
}

// New initialises the main state of the application
func New(startupOptions *StartupOptions) (App, error) {
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
	var tlsConfig *tls.Config
	if startupOptions.Redis.TLS {
		tlsConfig = &tls.Config{}
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:      startupOptions.Redis.Address,
		Password:  startupOptions.Redis.Password,
		TLSConfig: tlsConfig,
		DB:        startupOptions.Redis.Database,
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

	httpClient := http.DefaultClient
	integrationClient := integrationclient.New(httpClient, startupOptions.AuthenticationURL, startupOptions.ResourcesAllowedURL, startupOptions.UsageURL)

	return &AppImplementation{
		BlobStore:         storeAccess,
		JobDispatcher:     jobDispatcher,
		Stream:            streamClient,
		KeyValueStore:     keyValueStore,
		HTTP:              httpClient,
		integrationClient: integrationClient,
	}, nil
}

// CreateRun creates a run
func (app *AppImplementation) CreateRun(options protocol.CreateRunOptions) (protocol.Run, error) {
	// Generate run ID using uuid
	runID := uuid.NewV4().String()

	err := app.integrationClient.Authenticate(runID, options.APIKey)
	if err != nil {
		if errors.Is(err, integrationclient.ErrNotAllowed) {
			err = ErrNotAllowed
		}
		return protocol.Run{}, err
	}

	// Register in the key-value store that the run has been created
	// TODO: Error if the key already exists - probably want to use redis SETNX
	err = app.newCreatedKey(runID).set(true)
	return protocol.Run{ID: runID}, err
}

// GetApp downloads the tar & gzipped application code
func (app *AppImplementation) GetApp(runID string) (io.Reader, error) {
	return app.getBlobStoreData(runID, filenameApp)
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
		//nolint:errcheck // ignore error while handling an error
		//skipcq: GSC-G104
		os.Remove(tmpfile.Name())
		return tmpfile, fmt.Errorf("%w: %v", ErrArchiveFormat, err)
	}

	// Go back to the beginning of the temporary file
	_, err = tmpfile.Seek(0, io.SeekStart)
	if err != nil {
		//nolint:errcheck // ignore error while handling an error
		//skipcq: GSC-G104
		os.Remove(tmpfile.Name())
		return tmpfile, err
	}
	return tmpfile, nil
}

// PutApp uploads the tar & gzipped application code
func (app *AppImplementation) PutApp(runID string, reader io.Reader, objectSize int64) error {
	tmpfile, err := app.validateArchiveToTempFile(reader)
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())

	// Now upload the contents of the temporary file
	return app.putBlobStoreData(tmpfile, objectSize, runID, filenameApp)
}

// GetCache downloads the tar & gzipped build cache
func (app *AppImplementation) GetCache(runID string) (io.Reader, error) {
	return app.getBlobStoreData(runID, filenameCache)
}

// PutCache uploads the tar & gzipped build cache
func (app *AppImplementation) PutCache(runID string, reader io.Reader, objectSize int64) error {
	tmpfile, err := app.validateArchiveToTempFile(reader)
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())

	return app.putBlobStoreData(tmpfile, objectSize, runID, filenameCache)
}

// GetOutput downloads the scraper output
func (app *AppImplementation) GetOutput(runID string) (io.Reader, error) {
	return app.getBlobStoreData(runID, filenameOutput)
}

// PutOutput uploads the scraper output
func (app *AppImplementation) PutOutput(runID string, reader io.Reader, objectSize int64) error {
	return app.putBlobStoreData(reader, objectSize, runID, filenameOutput)
}

func (app *AppImplementation) setExitDataStage(runID string, stage string, value protocol.ExitDataStage) error {
	return app.newExitDataKey(runID, stage).set(value)
}

func (app *AppImplementation) getExitDataStage(runID string, stage string) (*protocol.ExitDataStage, error) {
	var exitDataStage protocol.ExitDataStage

	err := app.newExitDataKey(runID, stage).get(&exitDataStage)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &exitDataStage, nil
}

// GetExitData downloads the exit data
func (app *AppImplementation) GetExitData(runID string) (protocol.ExitData, error) {
	var exitData protocol.ExitData
	build, err := app.getExitDataStage(runID, "build")
	if err != nil {
		return exitData, err
	}
	exitData.Build = build
	execute, err := app.getExitDataStage(runID, "execute")
	if err != nil {
		return exitData, err
	}
	exitData.Execute = execute
	var exitDataFinished bool
	err = app.newExitDataFinishedKey(runID).get(&exitDataFinished)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return exitData, err
	}
	exitData.Finished = exitDataFinished
	return exitData, nil
}

// StartRun starts the run
func (app *AppImplementation) StartRun(runID string, dockerImage string, options protocol.StartRunOptions) error {
	// First check that the app exists
	_, err := app.GetApp(runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrAppNotAvailable
		}
		return err
	}

	err = app.integrationClient.ResourcesAllowed(runID, options.Memory, options.MaxRunTime)
	if err != nil {
		if errors.Is(err, integrationclient.ErrNotAllowed) {
			err = ErrNotAllowed
		}
		return err
	}

	err = app.newCallbackKey(runID).set(options.Callback.URL)
	if err != nil {
		return err
	}
	// We also store the amount of memory allocated in the key-value store because
	// we want to know this later (for reporting usage) but we don't want to have to
	// query k8s
	err = app.newMemoryKey(runID).set(options.Memory)
	if err != nil {
		return err
	}

	// Convert environment variable values to a single string that can be passed
	// as a flag to wrapper
	records := make([]string, 0, len(options.Env)>>1)
	for _, v := range options.Env {
		records = append(records, v.Name+"="+v.Value)
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
		runID,
		"--output", options.Output,
	}
	if envString != "" {
		command = append(command, "--env", envString)
	}
	return app.JobDispatcher.Create(runID, dockerImage, command, options.MaxRunTime, options.Memory)
}

// Events is an iterator to retrieve events from a stream
type Events struct {
	app    *AppImplementation
	runID  string
	lastID string
	more   bool
}

// GetEvents returns an iterator to get at all the events.
// Use "0" for lastId to start at the beginning of the stream. Otherwise use the id of the last
// seen event to restart the stream from that point. Don't try to restart the stream from the
// last event, otherwise More() will just wait around forever.
func (app *AppImplementation) GetEvents(runID string, lastID string) EventIterator {
	return &Events{app: app, runID: runID, lastID: lastID, more: true}
}

// More checks whether there are more events available. If true you can then call Next()
func (events *Events) More() bool {
	return events.more
}

// Next returns the next event
func (events *Events) Next() (e protocol.Event, err error) {
	e, err = events.app.Stream.Get(events.runID, events.lastID)
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
func (app *AppImplementation) CreateEvent(runID string, event protocol.Event) error {
	// TODO: Use something like runID-events instead for the stream name
	event, err := app.Stream.Add(runID, event)
	if err != nil {
		return err
	}
	// If this is a start, finish or last event do some extra special handling
	switch f := event.Data.(type) {
	case protocol.FirstData:
		// Record the time that this was started
		err = app.newFirstTimeKey(runID).set(event.Time)
		if err != nil {
			return err
		}
	case protocol.FinishData:
		err = app.setExitDataStage(runID, f.Stage, f.ExitData)
		if err != nil {
			return err
		}
		err := app.integrationClient.ReportNetworkUsage(runID, f.Stage, f.ExitData.Usage.NetworkIn, f.ExitData.Usage.NetworkOut)
		if err != nil {
			return err
		}
	case protocol.LastData:
		err = app.newExitDataFinishedKey(runID).set(true)
		if err != nil {
			return err
		}
		// Now determine how long the container was alive for and report back
		var firstTime time.Time
		err = app.newFirstTimeKey(runID).get(&firstTime)
		if err != nil {
			return err
		}
		// We also need the amount of memory allocated during the run
		var memory int64
		err = app.newMemoryKey(runID).get(&memory)
		if err != nil {
			return err
		}

		duration := event.Time.Sub(firstTime)
		// TODO: Convert memory to uint64?
		err = app.integrationClient.ReportMemoryUsage(runID, uint64(memory), duration)
		if err != nil {
			return err
		}
	}
	// We're intentionally doing the callback synchronously with the create event API call.
	// This way we can ensure that events within a run maintain their ordering.
	return app.postCallbackEvent(runID, event)
}

// DeleteRun deletes the run. Should be the last thing called
// TODO: If one delete operation fails the rest should still be attempted
func (app *AppImplementation) DeleteRun(runID string) error {
	err := app.JobDispatcher.Delete(runID)
	if err != nil {
		return err
	}
	err = app.deleteBlobStoreData(runID, filenameApp)
	if err != nil {
		return err
	}
	err = app.deleteBlobStoreData(runID, filenameOutput)
	if err != nil {
		return err
	}
	err = app.deleteBlobStoreData(runID, filenameCache)
	if err != nil {
		return err
	}
	err = app.deleteAllKeys(runID)
	if err != nil {
		return err
	}
	return app.Stream.Delete(runID)
}

func (app *AppImplementation) IsRunCreated(runID string) (bool, error) {
	var v bool
	err := app.newCreatedKey(runID).get(&v)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return v, err
	}
	return v, nil
}

// Submits a callback using a POST with the body set to the data serialised as JSON
// Returns an approximation of the number of bytes written
func (app *AppImplementation) postCallback(url string, data interface{}) (int, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return 0, err
	}

	client := retryablehttp.NewClient()
	client.HTTPClient = app.HTTP

	resp, err := client.Post(url, "application/json", b)
	if err != nil {
		// If we're erroring just assume that no data was sent
		return 0, err
	}
	defer resp.Body.Close()

	// TODO: We're ignoring the automated retries on the callbacks here in figuring out amount of traffic
	if resp.StatusCode != http.StatusOK {
		return len(b), errors.New("callback: " + resp.Status)
	}
	return len(b), nil
}

func (app *AppImplementation) postCallbackEvent(runID string, event protocol.Event) error {
	var callbackURL string
	err := app.newCallbackKey(runID).get(&callbackURL)
	if err != nil {
		return err
	}

	// Only do the callback if there's a sensible URL
	if callbackURL != "" {
		size, err := app.postCallback(callbackURL, event)
		// Record amount written even if there was an error
		if size > 0 {
			err = app.integrationClient.ReportNetworkUsage(runID, "callback", 0, uint64(size))
			if err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (app *AppImplementation) ReportAPINetworkUsage(runID string, in uint64, out uint64) error {
	return app.integrationClient.ReportNetworkUsage(runID, "api", in, out)
}
