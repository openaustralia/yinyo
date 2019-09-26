package client

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type createRunResult struct {
	RunName  string `json:"run_name"`
	RunToken string `json:"run_token"`
}

// TODO: Handle server being at a different URL
func createRun(namePrefix string) (createRunResult, error) {
	var result createRunResult

	uri := "http://localhost:8080/runs"
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
