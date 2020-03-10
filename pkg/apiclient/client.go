package apiclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/openaustralia/yinyo/pkg/protocol"
)

// Run is what you get when you create a run and what you need to update it
type Run struct {
	protocol.Run
	// Ignore this field when converting from/to json
	Client *Client
}

// RunInterface is the interface to interact with existing runs
type RunInterface interface {
	GetID() string
	GetApp() (io.ReadCloser, error)
	GetCache() (io.ReadCloser, error)
	GetOutput() (io.ReadCloser, error)
	GetExitData() (exitData protocol.ExitData, err error)
	PutApp(data io.Reader) error
	PutCache(data io.Reader) error
	PutOutput(data io.Reader) error
	Start(options *protocol.StartRunOptions) error
	GetEvents(lastID string) (*EventIterator, error)
	CreateEvent(event protocol.Event) (int, error)
	Delete() error
	// The following methods operate on to top of the lower level methods above
	// TODO: Should the following methods be in a separate interface?
	GetAppToDirectory(dir string) error
	PutAppFromDirectory(dir string, ignorePaths []string) error
	GetCacheToFile(path string) error
	GetCacheToDirectory(dir string) error
	PutCacheFromDirectory(dir string) error
	GetOutputToFile(path string) error
	PutOutputFromFile(path string) error
	CreateStartEvent(stage string) (int, error)
	CreateFinishEvent(stage string, exitData protocol.ExitDataStage) (int, error)
	CreateLogEvent(stage string, stream string, text string) (int, error)
	CreateNetworkEvent(in uint64, out uint64) (int, error)
	CreateLastEvent() (int, error)
}

// Client is used to access the API
type Client struct {
	URL        string
	HTTPClient *http.Client
}

// New configures a new Client
func New(url string) *Client {
	return &Client{
		URL:        url,
		HTTPClient: http.DefaultClient,
	}
}

func checkOK(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return errors.New(resp.Status)
}

// IsNotFound checks whether a particular error message corresponds to a 404
func IsNotFound(err error) bool {
	// TODO: Don't want to depend on a hardcoded string here
	return (err.Error() == "404 Not Found")
}

// IsUnauthorized checks whether a particular error message corresponds to a 401
func IsUnauthorized(err error) bool {
	// TODO: Don't want to depend on a hardcoded string here
	return (err.Error() == "401 Unauthorized")
}

func checkContentType(resp *http.Response, expected string) error {
	ct := resp.Header["Content-Type"]
	if len(ct) == 1 && ct[0] == expected {
		return nil
	}
	return errors.New("unexpected content type")
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
func (client *Client) CreateRun(options protocol.CreateRunOptions) (RunInterface, error) {
	run := &Run{Client: client}

	uri := client.URL + "/runs"
	if options.APIKey != "" {
		v := url.Values{}
		v.Add("api_key", options.APIKey)
		v.Add("callback_url", options.CallbackURL)
		uri = uri + "?" + v.Encode()
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

// GetID returns the name of the run
func (run *Run) GetID() string {
	return run.ID
}

// Make an API call for a particular run.
func (run *Run) request(method string, path string, body io.Reader) (*http.Response, error) {
	url := run.Client.URL + fmt.Sprintf("/runs/%s", run.ID) + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	return run.Client.HTTPClient.Do(req)
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

// GetOutput downloads the output of the run. Could be any file in any format.
func (run *Run) GetOutput() (io.ReadCloser, error) {
	resp, err := run.request("GET", "/output", nil)
	if err != nil {
		return nil, err
	}
	if err = checkOK(resp); err != nil {
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

// PutOutput uploads the output of the scraper
func (run *Run) PutOutput(data io.Reader) error {
	resp, err := run.request("PUT", "/output", data)
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

// Start starts a run that has earlier been created
// TODO: Add setting of environment variables
func (run *Run) Start(options *protocol.StartRunOptions) error {
	// TODO: Switch this over to using a json encoder
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

// EventIterator is a stream of events
type EventIterator struct {
	decoder *json.Decoder
}

// More checks whether another event is available
func (iterator *EventIterator) More() bool {
	return iterator.decoder.More()
}

// Next returns the next event
func (iterator *EventIterator) Next() (event protocol.Event, err error) {
	err = iterator.decoder.Decode(&event)
	return
}

// GetEvents returns a stream of events from the API
// If lastID is empty ("") then the stream starts from the beginning. Otherwise
// it starts from the first event after the one with the given ID.
func (run *Run) GetEvents(lastID string) (*EventIterator, error) {
	q := url.Values{}
	q.Add("last_id", lastID)
	resp, err := run.request("GET", "/events?"+q.Encode(), nil)
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

// CreateEvent sends an event and returns an approximation of the number of bytes sent
func (run *Run) CreateEvent(event protocol.Event) (int, error) {
	b, err := json.Marshal(event)
	if err != nil {
		return 0, err
	}
	resp, err := run.request("POST", "/events", bytes.NewReader(b))
	if err != nil {
		return 0, err
	}
	return len(b), checkOK(resp)
}

// GetExitData gets data about resource usage after everything has finished
func (run *Run) GetExitData() (exitData protocol.ExitData, err error) {
	resp, err := run.request("GET", "/exit-data", nil)
	if err != nil {
		return
	}
	if err = checkOK(resp); err != nil {
		return
	}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&exitData)
	return
}

// Delete cleans up after a run is complete
func (run *Run) Delete() error {
	resp, err := run.request("DELETE", "", nil)
	if err != nil {
		return err
	}
	return checkOK(resp)
}
