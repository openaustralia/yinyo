package clayclient

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// Run is what you get when you create a run and what you need to update it
type Run struct {
	Name  string `json:"run_name"`
	Token string `json:"run_token"`
	// Ignore this field when converting from/to json
	Client *Client
}

// Client is used to access the API
type Client struct {
	URL        string
	HTTPClient *http.Client
}

// New configures a new Client
func New(URL string) *Client {
	return &Client{
		URL:        URL,
		HTTPClient: http.DefaultClient,
	}
}

func checkOK(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return errors.New(resp.Status)
}

func checkContentType(resp *http.Response, expected string) error {
	ct := resp.Header["Content-Type"]
	if len(ct) == 1 && ct[0] == expected {
		return nil
	}
	return errors.New("Unexpected content type")
}

// Hello does a simple ping type request to the API
func (client *Client) Hello() (string, error) {
	req, err := http.NewRequest("GET", client.URL, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	if err = checkOK(resp); err != nil {
		return "", err
	}
	if err = checkContentType(resp, "text/plain; charset=utf-8"); err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CreateRun is the first thing called. It creates a run
func (client *Client) CreateRun(namePrefix string) (Run, error) {
	run := Run{Client: client}

	uri := client.URL + "/runs"
	if namePrefix != "" {
		params := url.Values{}
		params.Add("name_prefix", namePrefix)
		uri += "?" + params.Encode()
	}
	req, err := http.NewRequest("POST", uri, nil)
	if err != nil {
		return run, err
	}

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return run, err
	}
	if err = checkOK(resp); err != nil {
		return run, err
	}
	if err = checkContentType(resp, "application/json"); err != nil {
		return run, err
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&run)
	return run, err
}

// Make an API call for a particular run. These requests are always authenticated
func (run *Run) request(method string, path string, body io.Reader) (*http.Response, error) {
	url := run.Client.URL + fmt.Sprintf("/runs/%s", run.Name) + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+run.Token)
	return run.Client.HTTPClient.Do(req)
}

func extractArchiveToDirectory(gzipTarContent io.ReadCloser, dir string) error {
	gzipReader, err := gzip.NewReader(gzipTarContent)
	if err != nil {
		return err
	}
	tarReader := tar.NewReader(gzipReader)
	for {
		file, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}
		switch file.Typeflag {
		case tar.TypeDir:
			// TODO: Extract variable
			err := os.Mkdir(filepath.Join(dir, file.Name), 0755)
			if err != nil {
				return err
			}
		case tar.TypeReg:
			f, err := os.OpenFile(
				filepath.Join(dir, file.Name),
				os.O_RDWR|os.O_CREATE|os.O_TRUNC,
				file.FileInfo().Mode(),
			)
			if err != nil {
				return err
			}
			io.Copy(f, tarReader)
			f.Close()
		case tar.TypeSymlink:
			newname := filepath.Join(dir, file.Name)
			oldname := filepath.Join(filepath.Dir(newname), file.Linkname)
			err = os.Symlink(oldname, newname)
			if err != nil {
				return err
			}
		default:
			return errors.New("Unexpected type in tar")
		}
	}
	return nil
}

// createArchiveFromDirectory creates an archive from a directory on the filesystem
func createArchiveFromDirectory(dir string) (io.Reader, error) {
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		// Now put the contents of this file into the tar file
		tarWriter.WriteHeader(&tar.Header{
			Name: file.Name(),
			Size: file.Size(),
			Mode: 0600,
		})
		f, err := os.Open(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}
		io.Copy(tarWriter, f)
	}
	// TODO: This should always get called
	tarWriter.Close()
	gzipWriter.Close()
	return &buffer, nil
}

// GetAppToDirectory downloads the scraper code into a pre-existing directory on the filesystem
func (run *Run) GetAppToDirectory(dir string) error {
	app, err := run.GetApp()
	defer app.Close()
	if err != nil {
		return err
	}
	return extractArchiveToDirectory(app, dir)
}

// PutAppFromDirectory uploads the scraper code from a directory on the filesystem
func (run *Run) PutAppFromDirectory(dir string) error {
	r, err := createArchiveFromDirectory(dir)
	if err != nil {
		return err
	}
	return run.PutApp(r)
}

// GetCacheToDirectory downloads the cache into a pre-existing directory on the filesystem
func (run *Run) GetCacheToDirectory(dir string) error {
	app, err := run.GetCache()
	defer app.Close()
	if err != nil {
		return err
	}
	return extractArchiveToDirectory(app, dir)
}

// PutCacheFromDirectory uploads the cache from a directory on the filesystem
func (run *Run) PutCacheFromDirectory(dir string) error {
	r, err := createArchiveFromDirectory(dir)
	if err != nil {
		return err
	}
	return run.PutCache(r)
}

// GetApp downloads the tarred & gzipped scraper code
func (run *Run) GetApp() (io.ReadCloser, error) {
	resp, err := run.request("GET", "/app", nil)
	if err != nil {
		return nil, err
	}
	if err = checkOK(resp); err != nil {
		return nil, err
	}
	if err = checkContentType(resp, "application/gzip"); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// PutApp uploads the tarred & gzipped scraper code
func (run *Run) PutApp(appData io.Reader) error {
	resp, err := run.request("PUT", "/app", appData)
	if err != nil {
		return err
	}
	return checkOK(resp)
}

// PutCache uploads the tarred & gzipped build cache
func (run *Run) PutCache(data io.Reader) error {
	resp, err := run.request("PUT", "/cache", data)
	if err != nil {
		return err
	}
	return checkOK(resp)
}

// GetCache downloads the tarred & gzipped build cache
func (run *Run) GetCache() (io.ReadCloser, error) {
	resp, err := run.request("GET", "/cache", nil)
	if err != nil {
		return nil, err
	}
	if err = checkOK(resp); err != nil {
		return nil, err
	}
	if err = checkContentType(resp, "application/gzip"); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// StartRunOptions are options that can be used when starting a run
type StartRunOptions struct {
	Output string
}

// Start starts a run that has earlier been created
// TODO: Add setting of environment variables
func (run *Run) Start(options *StartRunOptions) error {
	b, err := json.Marshal(options)
	if err != nil {
		return err
	}
	resp, err := run.request("POST", "/start", bytes.NewReader(b))
	if err != nil {
		return err
	}
	return checkOK(resp)
}

type eventRaw struct {
	Stage  string `json:"stage"`
	Type   string `json:"type"`
	Stream string `json:"stream,omitempty"`
	Text   string `json:"text,omitempty"`
}

// Event is the interface for all event types
type Event interface {
}

// StartEvent represents the start of a build or run
type StartEvent struct {
	Stage string
}

// FinishEvent represent the completion of a build or run
type FinishEvent struct {
	Stage string
}

// LogEvent is the output of some text from the build or run of a scraper
type LogEvent struct {
	Stage  string
	Stream string
	Text   string
}

// EventIterator is a stream of events
type EventIterator struct {
	decoder *json.Decoder
}

// More checks whether another event is available
func (iterator *EventIterator) More() bool {
	return iterator.decoder.More()
}

// Next returns the next event
func (iterator *EventIterator) Next() (Event, error) {
	var eventRaw eventRaw
	err := iterator.decoder.Decode(&eventRaw)
	if err != nil {
		return nil, err
	}
	if eventRaw.Type == "started" {
		return StartEvent{Stage: eventRaw.Stage}, nil
	} else if eventRaw.Type == "finished" {
		return FinishEvent{Stage: eventRaw.Stage}, nil
	} else if eventRaw.Type == "log" {
		return LogEvent{Stage: eventRaw.Stage, Stream: eventRaw.Stream, Text: eventRaw.Text}, nil
	}
	return nil, errors.New("Unexpected type")
}

// GetEvents returns a stream of events from the API
func (run *Run) GetEvents() (*EventIterator, error) {
	resp, err := run.request("GET", "/events", nil)
	if err != nil {
		return nil, err
	}
	if err = checkOK(resp); err != nil {
		return nil, err
	}
	if err = checkContentType(resp, "application/ld+json"); err != nil {
		return nil, err
	}
	return &EventIterator{decoder: json.NewDecoder(resp.Body)}, nil
}

// CreateStartEvent sends an event signalling the start of a "build" or "run"
func (run *Run) CreateStartEvent(stage string) error {
	event := eventRaw{Type: "start", Stage: stage}
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	resp, err := run.request("POST", "/events", bytes.NewReader(b))
	if err != nil {
		return err
	}
	return checkOK(resp)
}

// CreateFinishEvent sends an event signalling the start of a "build" or "run"
func (run *Run) CreateFinishEvent(stage string) error {
	event := eventRaw{Type: "finish", Stage: stage}
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	resp, err := run.request("POST", "/events", bytes.NewReader(b))
	if err != nil {
		return err
	}
	return checkOK(resp)
}

// CreateLogEvent sends an event with a single log line
// TODO: Factor out common code with CreateStartEvent
func (run *Run) CreateLogEvent(stage string, stream string, text string) error {
	event := eventRaw{Type: "log", Stage: stage, Stream: stream, Text: text}
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}
	resp, err := run.request("POST", "/events", bytes.NewReader(b))
	if err != nil {
		return err
	}
	return checkOK(resp)
}

// Delete cleans up after a run is complete
func (run *Run) Delete() error {
	resp, err := run.request("DELETE", "", nil)
	if err != nil {
		return err
	}
	return checkOK(resp)
}
