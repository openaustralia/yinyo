package client

import (
	"encoding/json"
	"fmt"
	"io"
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

func (client *Client) Hello() (*http.Response, error) {
	req, err := http.NewRequest("GET", client.URL, nil)
	if err != nil {
		return nil, err
	}
	return client.HttpClient.Do(req)
}

func (client *Client) CreateRun(namePrefix string) (Run, error) {
	var result Run

	uri := client.URL + "/runs"
	if namePrefix != "" {
		params := url.Values{}
		params.Add("name_prefix", namePrefix)
		uri += "?" + params.Encode()
	}
	req, err := http.NewRequest("POST", uri, nil)
	if err != nil {
		return result, err
	}

	resp, err := client.HttpClient.Do(req)
	if err != nil {
		return result, err
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&result)
	return result, err
}

func (client *Client) PutApp(run Run, appData io.Reader) (*http.Response, error) {
	url := client.URL + fmt.Sprintf("/runs/%s/app", run.Name)
	req, err := http.NewRequest("PUT", url, appData)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+run.Token)
	return client.HttpClient.Do(req)
}

func (client *Client) GetApp(run Run) (*http.Response, error) {
	url := client.URL + fmt.Sprintf("/runs/%s/app", run.Name)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+run.Token)
	return client.HttpClient.Do(req)
}
