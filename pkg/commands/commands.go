package commands

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/openaustralia/yinyo/pkg/archive"
	"github.com/openaustralia/yinyo/pkg/blobstore"
	"github.com/openaustralia/yinyo/pkg/jobdispatcher"
	"github.com/openaustralia/yinyo/pkg/keyvaluestore"
	"github.com/openaustralia/yinyo/pkg/protocol"
	"github.com/openaustralia/yinyo/pkg/stream"
)

const dockerImage = "openaustralia/yinyo-scraper:v1"
const runBinary = "/bin/wrapper"

// App is the interface for the operations of the server
type App interface {
	CreateRun(apiKey string) (protocol.Run, error)
	DeleteRun(runID string) error
	StartRun(runID string, output string, env map[string]string, callbackURL string, maxRunTime int64) error
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
	RecordAPINetworkUsage(runID string, external bool, in int64, out int64) error
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
	AuthenticationURL string
	UsageURL          string
}

// StartupOptions are the options available when initialising the application
type StartupOptions struct {
	Minio             MinioOptions
	Redis             RedisOptions
	AuthenticationURL string
	UsageURL          string
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
		BlobStore:         storeAccess,
		JobDispatcher:     jobDispatcher,
		Stream:            streamClient,
		KeyValueStore:     keyValueStore,
		HTTP:              http.DefaultClient,
		AuthenticationURL: startupOptions.AuthenticationURL,
		UsageURL:          startupOptions.UsageURL,
	}, nil
}

type authenticationResponse struct {
	Allowed bool   `json:"allowed"`
	Message string `json:"message"`
}

// CreateRun creates a run
func (app *AppImplementation) CreateRun(apiKey string) (protocol.Run, error) {
	if app.AuthenticationURL != "" {
		v := url.Values{}
		v.Add("api_key", apiKey)
		url := app.AuthenticationURL + "?" + v.Encode()
		log.Printf("Making an authentication request to %v", url)

		resp, err := app.HTTP.Post(url, "application/json", nil)
		if err != nil {
			return protocol.Run{}, err
		}
		if resp.StatusCode != http.StatusOK {
			return protocol.Run{}, fmt.Errorf("Response %v from POST to authentication url %v", resp.StatusCode, url)
		}
		// TODO: Check the actual response
		dec := json.NewDecoder(resp.Body)
		var response authenticationResponse
		err = dec.Decode(&response)
		if err != nil {
			return protocol.Run{}, err
		}
		if !response.Allowed {
			// TODO: Send the message back to the user
			return protocol.Run{}, ErrNotAllowed
		}

		// TODO: Do we want to do something with response.Message if allowed?

		err = resp.Body.Close()
		if err != nil {
			return protocol.Run{}, err
		}
	}
	// Generate run ID using uuid
	id := uuid.NewV4().String()

	// Register in the key-value store that the run has been created
	// TODO: Error if the key already exists - probably want to use redis SETNX
	err := app.newCreatedKey(id).set("true")
	return protocol.Run{ID: id}, err
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

// GetExitData downloads the exit data
func (app *AppImplementation) GetExitData(runID string) (protocol.ExitData, error) {
	var exitData protocol.ExitData
	build, err := app.newExitDataKey(runID, "build").get()
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return exitData, err
		}
	} else {
		var exitDataBuild protocol.ExitDataStage
		err = json.Unmarshal([]byte(build), &exitDataBuild)
		if err != nil {
			return exitData, err
		}
		exitData.Build = &exitDataBuild
	}
	run, err := app.newExitDataKey(runID, "run").get()
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return exitData, err
		}
	} else {
		var exitDataRun protocol.ExitDataStage
		err = json.Unmarshal([]byte(run), &exitDataRun)
		if err != nil {
			return exitData, err
		}
		exitData.Run = &exitDataRun
	}
	finished, err := app.newExitDataFinishedKey(runID).get()
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return exitData, err
		}
	} else {
		var exitDataFinished bool
		err = json.Unmarshal([]byte(finished), &exitDataFinished)
		if err != nil {
			return exitData, err
		}
		exitData.Finished = exitDataFinished
	}
	apiNetworkIn, err := app.newExitDataAPINetworkInKey(runID).getAsInt()
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return exitData, err
		}
	} else {
		exitData.API.NetworkIn = uint64(apiNetworkIn)
	}
	apiNetworkOut, err := app.newExitDataAPINetworkOutKey(runID).getAsInt()
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return exitData, err
		}
	} else {
		exitData.API.NetworkOut = uint64(apiNetworkOut)
	}
	return exitData, nil
}

// StartRun starts the run
func (app *AppImplementation) StartRun(
	runID string, output string, env map[string]string, callbackURL string, maxRunTime int64) error {
	// First check that the app exists
	_, err := app.GetApp(runID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrAppNotAvailable
		}
		return err
	}

	err = app.newCallbackKey(runID).set(callbackURL)
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
		runID,
		"--output", output,
	}
	if envString != "" {
		command = append(command, "--env", envString)
	}
	return app.JobDispatcher.Create(runID, dockerImage, command, maxRunTime)
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
	// If this is a finish event or a last event do some extra special handling
	switch f := event.Data.(type) {
	case protocol.FinishData:
		exitDataBytes, err := json.Marshal(f.ExitData)
		if err != nil {
			return err
		}
		err = app.newExitDataKey(runID, f.Stage).set(string(exitDataBytes))
		if err != nil {
			return err
		}
	case protocol.LastData:
		err = app.newExitDataFinishedKey(runID).set("true")
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
	v, err := app.newCreatedKey(runID).get()
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return v == "true", nil
}

func (app *AppImplementation) postCallbackEvent(runID string, event protocol.Event) error {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	err := enc.Encode(event)
	if err != nil {
		return err
	}

	callbackURL, err := app.newCallbackKey(runID).get()
	if err != nil {
		return err
	}

	// Only do the callback if there's a sensible URL
	if callbackURL != "" {
		// Capture the size of the outgoing request now (before it's read)
		out := b.Len()

		client := retryablehttp.NewClient()
		client.HTTPClient = app.HTTP

		resp, err := client.Post(callbackURL, "application/json", &b)
		if err != nil {
			// TODO: In case of an error we've probably still sent traffic. Record this.
			return err
		}
		defer resp.Body.Close()

		// TODO: We're ignoring the automated retries on the callbacks here in figuring out amount of traffic
		err = app.RecordAPINetworkUsage(runID, true, 0, int64(out))
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return errors.New("callback: " + resp.Status)
		}
	}
	return nil
}

func (app *AppImplementation) RecordAPINetworkUsage(runID string, external bool, in int64, out int64) error {
	// We only record traffic that is going out or coming in via the public internet
	if external {
		_, err := app.newExitDataAPINetworkInKey(runID).increment(in)
		if err != nil {
			return err
		}
		_, err = app.newExitDataAPINetworkOutKey(runID).increment(out)
		if err != nil {
			return err
		}
	}
	return nil
}
