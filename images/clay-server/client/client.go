package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type createRunResult struct {
	RunName  string `json:"run_name"`
	RunToken string `json:"run_token"`
}

type Client struct {
	URL string
}

func NewClient(URL string) Client {
	return Client{URL: URL}
}

// TODO: Handle server being at a different URL
func (client *Client) CreateRun(namePrefix string) (createRunResult, error) {
	var result createRunResult

	uri := client.URL + "/runs"
	if namePrefix != "" {
		params := url.Values{}
		params.Add("name_prefix", namePrefix)
		uri += "?" + params.Encode()
	}
	resp, err := http.Post(uri, "", nil)
	if err != nil {
		return result, err
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&result)
	return result, err
}

func (client *Client) PutApp(run createRunResult, appData io.Reader) (*http.Response, error) {
	url := client.URL + fmt.Sprintf("/runs/%s/app", run.RunName)
	req, err := http.NewRequest("PUT", url, appData)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+run.RunToken)
	return http.DefaultClient.Do(req)
}

func (client *Client) GetApp(run createRunResult) (*http.Response, error) {
	url := client.URL + fmt.Sprintf("/runs/%s/app", run.RunName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+run.RunToken)
	return http.DefaultClient.Do(req)
}
