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

type Client struct {
	URL        string
	HttpClient *http.Client
}

func NewClient(URL string) Client {
	return Client{
		URL:        URL,
		HttpClient: http.DefaultClient,
	}
}

func (client *Client) helloRaw() (*http.Response, error) {
	req, err := http.NewRequest("GET", client.URL, nil)
	if err != nil {
		return nil, err
	}
	return client.HttpClient.Do(req)
}

func (client *Client) Hello() (string, error) {
	resp, err := client.helloRaw()
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(resp.Status)
	}
	ct := resp.Header["Content-Type"]
	if len(ct) != 1 || ct[0] != "text/plain; charset=utf-8" {
		return "", errors.New("Unexpected content type")
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

	return client.HttpClient.Do(req)
}

func (client *Client) CreateRun(namePrefix string) (Run, error) {
	var result Run

	resp, err := client.createRunRaw(namePrefix)
	if err != nil {
		return result, err
	}
	if resp.StatusCode != http.StatusOK {
		return result, errors.New(resp.Status)
	}
	ct := resp.Header["Content-Type"]
	if len(ct) != 1 || ct[0] != "application/json" {
		return result, errors.New("Unexpected content type")
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&result)
	return result, err
}

func (client *Client) putAppRaw(run Run, appData io.Reader) (*http.Response, error) {
	url := client.URL + fmt.Sprintf("/runs/%s/app", run.Name)
	req, err := http.NewRequest("PUT", url, appData)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+run.Token)
	return client.HttpClient.Do(req)
}

func (client *Client) PutApp(run Run, appData io.Reader) error {
	resp, err := client.putAppRaw(run, appData)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	return nil
}

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
	url := client.URL + fmt.Sprintf("/runs/%s/app", run.Name)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+run.Token)
	return client.HttpClient.Do(req)
}

func (client *Client) GetApp(run Run) (io.ReadCloser, error) {
	resp, err := client.getAppRaw(run)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	ct := resp.Header["Content-Type"]
	if len(ct) != 1 || ct[0] != "application/gzip" {
		return nil, errors.New("Unexpected content type")
	}
	return resp.Body, nil
}

func (client *Client) putCacheRaw(run Run, data io.Reader) (*http.Response, error) {
	url := client.URL + fmt.Sprintf("/runs/%s/cache", run.Name)
	req, err := http.NewRequest("PUT", url, data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+run.Token)
	return client.HttpClient.Do(req)
}

func (client *Client) PutCache(run Run, data io.Reader) error {
	resp, err := client.putCacheRaw(run, data)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	return nil
}

func (client *Client) getCacheRaw(run Run) (*http.Response, error) {
	url := client.URL + fmt.Sprintf("/runs/%s/cache", run.Name)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+run.Token)
	return client.HttpClient.Do(req)
}

func (client *Client) GetCache(run Run) (io.ReadCloser, error) {
	resp, err := client.getCacheRaw(run)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	ct := resp.Header["Content-Type"]
	if len(ct) != 1 || ct[0] != "application/gzip" {
		return nil, errors.New("Unexpected content type")
	}
	return resp.Body, nil
}

type StartRunOptions struct {
	Output string
}

// TODO: Add setting of environment variables
func (client *Client) startRunRaw(run Run, options *StartRunOptions) (*http.Response, error) {
	url := client.URL + fmt.Sprintf("/runs/%s/start", run.Name)
	b, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+run.Token)
	return client.HttpClient.Do(req)
}

func (client *Client) StartRun(run Run, options *StartRunOptions) error {
	resp, err := client.startRunRaw(run, options)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	return nil
}

type EventRaw struct {
	Stage  string
	Type   string
	Stream string
	Log    string // TODO: Rename Log to Text
}

type Event interface {
}

type StartEvent struct {
	Stage string
}

type FinishEvent struct {
	Stage string
}

type LogEvent struct {
	Stage  string
	Stream string
	Log    string // TODO: Rename Log to Text
}

type EventIterator struct {
	decoder *json.Decoder
}

func (iterator *EventIterator) More() bool {
	return iterator.decoder.More()
}

func (iterator *EventIterator) Next() (Event, error) {
	var eventRaw EventRaw
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

func (client *Client) GetEventsRaw(run Run) (*http.Response, error) {
	url := client.URL + fmt.Sprintf("/runs/%s/events", run.Name)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+run.Token)
	return client.HttpClient.Do(req)
}
