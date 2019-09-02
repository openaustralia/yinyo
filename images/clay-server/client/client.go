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
func createRun(scraperName string) (createRunResult, error) {
	var result createRunResult

	params := url.Values{}
	params.Add("scraper_name", scraperName)
	resp, err := http.Post("http://localhost:8080/runs?"+params.Encode(), "", nil)
	if err != nil {
		return result, err
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&result)
	return result, err
}
