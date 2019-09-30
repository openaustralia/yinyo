package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

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

func (client *Client) PutCacheRaw(run Run, data io.Reader) (*http.Response, error) {
	url := client.URL + fmt.Sprintf("/runs/%s/cache", run.Name)
	req, err := http.NewRequest("PUT", url, data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+run.Token)
	return client.HttpClient.Do(req)
}

func (client *Client) GetCacheRaw(run Run) (*http.Response, error) {
	url := client.URL + fmt.Sprintf("/runs/%s/cache", run.Name)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+run.Token)
	return client.HttpClient.Do(req)
}

type StartRunOptions struct {
	Output string
}

// TODO: Add setting of environment variables
func (client *Client) StartRunRaw(run Run, options *StartRunOptions) (*http.Response, error) {
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

func (client *Client) GetEventsRaw(run Run) (*http.Response, error) {
	url := client.URL + fmt.Sprintf("/runs/%s/events", run.Name)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+run.Token)
	return client.HttpClient.Do(req)
}
