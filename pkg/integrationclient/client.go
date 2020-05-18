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

var ErrNotAllowed = errors.New("not allowed")

func Authenticate(authenticationURL string, httpClient *http.Client, runID string, apiKey string) error {
	if authenticationURL != "" {
		v := url.Values{}
		v.Add("api_key", apiKey)
		v.Add("run_id", runID)
		url := authenticationURL + "?" + v.Encode()
		log.Printf("Making an authentication request to %v", url)

		resp, err := httpClient.Post(url, "application/json", nil)
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
			// TODO: Send the message back to the user
			return ErrNotAllowed
		}

		// TODO: Do we want to do something with response.Message if allowed?

		err = resp.Body.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func ResourcesAllowed(resourcesAllowedURL string, httpClient *http.Client, runID string, memory int64, maxRunTime int64) error {
	// Now check if the user is allowed the memory and the time
	// to start this run
	if resourcesAllowedURL != "" {
		v := url.Values{}
		v.Add("run_id", runID)
		v.Add("time", fmt.Sprint(maxRunTime))
		v.Add("memory", fmt.Sprint(memory))
		url := resourcesAllowedURL + "?" + v.Encode()
		log.Printf("Making a resources allowed request to %v", url)

		resp, err := httpClient.Post(url, "application/json", nil)
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
			// TODO: Send the message back to the user
			return ErrNotAllowed
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
// For the time being we're just logging stuff locally to show that it's happening
// TODO: Will need the reporting URL as well
func ReportNetworkUsage(runID string, source string, in uint64, out uint64) error {
	log.Printf("Network Usage: %v source: %v, in: %v, out: %v", runID, source, in, out)
	return nil
}

// ReportMemoryUsage lets an external system know about some memory usage
// For the time being we're just logging stuff locally to show that it's happening
// TODO: Will need the reporting URL as well
func ReportMemoryUsage(runID string, memory uint64, duration time.Duration) error {
	log.Printf("Memory Usage: %v memory: %v, duration: %v", runID, memory, duration)
	return nil
}
