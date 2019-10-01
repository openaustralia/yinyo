package client

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
}

// Client is used to access the API
type Client struct {
	URL        string
	HTTPClient *http.Client
}

// NewClient configures a new Client
func NewClient(URL string) Client {
	return Client{
		URL:        URL,
		HTTPClient: http.DefaultClient,
	}
}

func (client *Client) helloRaw() (*http.Response, error) {
	req, err := http.NewRequest("GET", client.URL, nil)
	if err != nil {
		return nil, err
	}
	return client.HTTPClient.Do(req)
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
	resp, err := client.helloRaw()
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

func (client *Client) createRunRaw(namePrefix string) (*http.Response, error) {
	uri := client.URL + "/runs"
	if namePrefix != "" {
		params := url.Values{}
		params.Add("name_prefix", namePrefix)
		uri += "?" + params.Encode()
	}
	req, err := http.NewRequest("POST", uri, nil)
	if err != nil {
		return nil, err
	}

	return client.HTTPClient.Do(req)
}

// CreateRun is the first thing called. It creates a run
func (client *Client) CreateRun(namePrefix string) (Run, error) {
	var result Run

	resp, err := client.createRunRaw(namePrefix)
	if err != nil {
		return result, err
	}
	if err = checkOK(resp); err != nil {
		return result, err
	}
	if err = checkContentType(resp, "application/json"); err != nil {
		return result, err
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&result)
	return result, err
}

func addAuthorization(req *http.Request, run Run) {
	req.Header.Set("Authorization", "Bearer "+run.Token)
}

// Make an API call for a particular run. These requests are always authenticated
func (client *Client) runRequest(run Run, method string, path string, body io.Reader) (*http.Response, error) {
	url := client.URL + fmt.Sprintf("/runs/%s/", run.Name) + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	addAuthorization(req, run)
	return client.HTTPClient.Do(req)
}

func (client *Client) putAppRaw(run Run, appData io.Reader) (*http.Response, error) {
	return client.runRequest(run, "PUT", "app", appData)
}

// PutApp uploads the tarred & gzipped scraper code
func (client *Client) PutApp(run Run, appData io.Reader) error {
	resp, err := client.putAppRaw(run, appData)
	if err != nil {
		return err
	}
	return checkOK(resp)
}

// PutAppFromDirectory uploads the scraper code from a directory on the filesystem
func (client *Client) PutAppFromDirectory(run Run, dir string) error {
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
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
			return err
		}
		io.Copy(tarWriter, f)
	}
	// TODO: This should always get called
	tarWriter.Close()
	gzipWriter.Close()

	return client.PutApp(run, &buffer)
}

func (client *Client) getAppRaw(run Run) (*http.Response, error) {
	return client.runRequest(run, "GET", "app", nil)
}

// GetApp downloads the tarred & gzipped scraper code
func (client *Client) GetApp(run Run) (io.ReadCloser, error) {
	resp, err := client.getAppRaw(run)
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

func (client *Client) putCacheRaw(run Run, data io.Reader) (*http.Response, error) {
	return client.runRequest(run, "PUT", "cache", data)
}

// PutCache uploads the tarred & gzipped build cache
func (client *Client) PutCache(run Run, data io.Reader) error {
	resp, err := client.putCacheRaw(run, data)
	if err != nil {
		return err
	}
	return checkOK(resp)
}

func (client *Client) getCacheRaw(run Run) (*http.Response, error) {
	return client.runRequest(run, "GET", "cache", nil)
}

// GetCache downloads the tarred & gzipped build cache
func (client *Client) GetCache(run Run) (io.ReadCloser, error) {
	resp, err := client.getCacheRaw(run)
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

// TODO: Add setting of environment variables
func (client *Client) startRunRaw(run Run, options *StartRunOptions) (*http.Response, error) {
	b, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}
	return client.runRequest(run, "POST", "start", bytes.NewReader(b))
}

// StartRun starts a run that has earlier been created
func (client *Client) StartRun(run Run, options *StartRunOptions) error {
	resp, err := client.startRunRaw(run, options)
	if err != nil {
		return err
	}
	return checkOK(resp)
}

type eventRaw struct {
	Stage  string
	Type   string
	Stream string
	Log    string // TODO: Rename Log to Text
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
	Log    string // TODO: Rename Log to Text
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
		return LogEvent{Stage: eventRaw.Stage, Stream: eventRaw.Stream, Log: eventRaw.Log}, nil
	}
	return nil, errors.New("Unexpected type")
}

func (client *Client) getEventsRaw(run Run) (*http.Response, error) {
	return client.runRequest(run, "GET", "events", nil)
}

// GetEvents returns a stream of events from the API
func (client *Client) GetEvents(run Run) (*EventIterator, error) {
	resp, err := client.getEventsRaw(run)
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
