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
	"os"

	"github.com/openaustralia/yinyo/pkg/archive"
	"github.com/openaustralia/yinyo/pkg/protocol"
)

// Run is what you get when you create a run and what you need to update it
type Run struct {
	protocol.Run
	// Ignore this field when converting from/to json
	Client *Client
}

type RunInterface interface {
	GetName() string
	GetToken() string
	GetApp() (io.ReadCloser, error)
	GetCache() (io.ReadCloser, error)
	GetOutput() (io.ReadCloser, error)
	GetExitData() (exitData protocol.ExitData, err error)
	PutApp(data io.Reader) error
	PutCache(data io.Reader) error
	PutOutput(data io.Reader) error
	PutExitData(exitData protocol.ExitData) error
	Start(options *protocol.StartRunOptions) error
	GetEvents(lastID string) (*EventIterator, error)
	CreateEvent(event protocol.Event) error
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
func (client *Client) CreateRun(namePrefix string) (RunInterface, error) {
	run := &Run{Client: client}

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

func (run *Run) GetName() string {
	return run.Name
}

func (run *Run) GetToken() string {
	return run.Token
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

// GetAppToDirectory downloads the scraper code into a pre-existing directory on the filesystem
func (run *Run) GetAppToDirectory(dir string) error {
	app, err := run.GetApp()
	if err != nil {
		return err
	}
	defer app.Close()
	return archive.ExtractToDirectory(app, dir)
}

// PutAppFromDirectory uploads the scraper code from a directory on the filesystem
// ignorePaths is a list of paths (relative to dir) that should be ignored and not uploaded
func (run *Run) PutAppFromDirectory(dir string, ignorePaths []string) error {
	r, err := archive.CreateFromDirectory(dir, ignorePaths)
	if err != nil {
		return err
	}
	return run.PutApp(r)
}

// GetCacheToDirectory downloads the cache into a pre-existing directory on the filesystem
func (run *Run) GetCacheToDirectory(dir string) error {
	app, err := run.GetCache()
	if err != nil {
		// If cache doesn't exist then do nothing
		if IsNotFound(err) {
			return nil
		}
		return err
	}
	defer app.Close()
	return archive.ExtractToDirectory(app, dir)
}

// PutCacheFromDirectory uploads the cache from a directory on the filesystem
func (run *Run) PutCacheFromDirectory(dir string) error {
	r, err := archive.CreateFromDirectory(dir, []string{})
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

// GetOutputToFile downloads the output of the run and saves it in a file which it
// will create or overwrite.
func (run *Run) GetOutputToFile(path string) error {
	output, err := run.GetOutput()
	if err != nil {
		return err
	}
	defer output.Close()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, output)
	return err
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

// PutOutputFromFile uploads the contents of a file as the output of the scraper
func (run *Run) PutOutputFromFile(path string) error {
	// TODO: Don't do a separate Stat and Open
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		return run.PutOutput(f)
	}
	// We get here if output file doesn't exist. In that case we just want
	// to happily carry on like nothing weird has happened
	return nil
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

// GetCacheToFile downloads the cache (as a tar & gzipped file) and saves it (without uncompressing it)
func (run *Run) GetCacheToFile(path string) error {
	cache, err := run.GetCache()
	if err != nil {
		return err
	}
	defer cache.Close()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, cache)
	return err
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

// CreateEvent sends an event
func (run *Run) CreateEvent(event protocol.Event) error {
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

// PutExitData uploads information about how things ran and how much resources were used
func (run *Run) PutExitData(exitData protocol.ExitData) error {
	b, err := json.Marshal(exitData)
	if err != nil {
		return err
	}
	resp, err := run.request("PUT", "/exit-data", bytes.NewReader(b))
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
