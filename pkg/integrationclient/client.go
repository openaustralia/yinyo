package integrationclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type authenticationResponse struct {
	Allowed bool   `json:"allowed"`
	Message string `json:"message"`
}

var ErrNotAllowed = errors.New("Not allowed")

type Client struct {
	httpClient          *http.Client
	authenticationURL   string
	resourcesAllowedURL string
	usageURL            string
}

func New(httpClient *http.Client, authenticationURL string, resourcesAllowedURL string, usageURL string) *Client {
	return &Client{httpClient: httpClient, authenticationURL: authenticationURL, resourcesAllowedURL: resourcesAllowedURL, usageURL: usageURL}
}

func (client *Client) Authenticate(runID string, apiKey string) error {
	if client.authenticationURL != "" {
		v := url.Values{}
		v.Add("api_key", apiKey)
		v.Add("run_id", runID)
		url := client.authenticationURL + "?" + v.Encode()
		log.Printf("Making an authentication request to %v", url)

		resp, err := client.httpClient.Post(url, "application/json", nil)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("Response %v from POST to authentication url %v", resp.StatusCode, url)
		}
		// TODO: Check the actual response
		dec := json.NewDecoder(resp.Body)
		var response authenticationResponse
		err = dec.Decode(&response)
		if err != nil {
			return err
		}
		if !response.Allowed {
			return fmt.Errorf("%w: %v", ErrNotAllowed, response.Message)
		}

		// TODO: Do we want to do something with response.Message if allowed?

		err = resp.Body.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (client *Client) ResourcesAllowed(runID string, memory int64, maxRunTime int64) error {
	// Now check if the user is allowed the memory and the time
	// to start this run
	if client.resourcesAllowedURL != "" {
		v := url.Values{}
		v.Add("run_id", runID)
		v.Add("time", fmt.Sprint(maxRunTime))
		v.Add("memory", fmt.Sprint(memory))
		url := client.resourcesAllowedURL + "?" + v.Encode()
		log.Printf("Making a resources allowed request to %v", url)

		resp, err := client.httpClient.Post(url, "application/json", nil)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("Response %v from POST to resources allowed url %v", resp.StatusCode, url)
		}
		// TODO: Check the actual response
		dec := json.NewDecoder(resp.Body)
		// We're using the same response as with authentication.
		// TODO: Rename this to something more generic
		var response authenticationResponse
		err = dec.Decode(&response)
		if err != nil {
			return err
		}
		if !response.Allowed {
			return fmt.Errorf("%w: %v", ErrNotAllowed, response.Message)
		}

		// TODO: Do we want to do something with response.Message if allowed?

		err = resp.Body.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// ReportNetworkUsage lets an external system know about some network usage
func (client *Client) ReportNetworkUsage(runID string, source string, in uint64, out uint64) error {
	// No need to report anything if there was no traffic
	if in > 0 || out > 0 {
		log.Printf("Network Usage: %v source: %v, in: %v, out: %v", runID, source, in, out)
		if client.usageURL != "" {
			v := url.Values{}
			v.Add("run_id", runID)
			v.Add("source", source)
			v.Add("in", fmt.Sprint(in))
			v.Add("out", fmt.Sprint(out))
			url := client.usageURL + "/network?" + v.Encode()
			log.Printf("Reporting network usage to %v", url)
			resp, err := client.httpClient.Post(url, "application/json", nil)
			if err != nil {
				return err
			}
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("Response %v from POST to %v", resp.StatusCode, url)
			}
			err = resp.Body.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ReportMemoryUsage lets an external system know about some memory usage
func (client *Client) ReportMemoryUsage(runID string, memory uint64, duration time.Duration) error {
	log.Printf("Memory Usage: %v memory: %v, duration: %v", runID, memory, duration)
	if client.usageURL != "" {
		v := url.Values{}
		v.Add("run_id", runID)
		v.Add("memory", fmt.Sprint(memory))
		v.Add("duration", fmt.Sprint(duration.Seconds()))
		url := client.usageURL + "/memory?" + v.Encode()
		log.Printf("Reporting memory usage to %v", url)
		resp, err := client.httpClient.Post(url, "application/json", nil)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("Response %v from POST to %v", resp.StatusCode, url)
		}
		err = resp.Body.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
