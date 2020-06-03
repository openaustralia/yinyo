package apiclient

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/openaustralia/yinyo/pkg/protocol"
)

const cacheName = ".yinyo-build-cache.tgz"

func uploadCacheIfExists(run RunInterface, cachePath string) error {
	file, err := os.Open(cachePath)
	if err != nil {
		// If the cache doesn't exist then skip the uploading bit
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		defer file.Close()
		err = run.PutCache(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func downloadOutput(run RunInterface, scraperDirectory string, outputFile string) error {
	// Get the run output
	if outputFile != "" {
		err := run.GetOutputToFile(filepath.Join(scraperDirectory, outputFile))
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				fmt.Printf("Warning: output file %v does not exist", outputFile)
			} else {
				return err
			}
		}
	}
	return nil
}

func reformatEnvironmentVariables(environment map[string]string) []protocol.EnvVariable {
	var envVariables []protocol.EnvVariable
	for k, v := range environment {
		// TODO: Fix this inefficient way
		envVariables = append(envVariables, protocol.EnvVariable{Name: k, Value: v})
	}
	return envVariables
}

// Simple is a super simple high level way of running a scraper that exists on the local file system
// giving you a local callback for every event (including logs). This is used by the command line client
// It makes a simple common use case a little simpler to implement
func Simple(scraperDirectory string, clientServerURL string, environment map[string]string,
	outputFile string, cache bool, callbackURL string, apiKey string, eventCallback func(event protocol.Event) error, showProgress bool) error {
	client := New(clientServerURL)
	// Create the run
	run, err := client.CreateRun(protocol.CreateRunOptions{APIKey: apiKey})
	if err != nil {
		return err
	}
	err = SimpleStart(run.GetID(), scraperDirectory, clientServerURL, environment, outputFile, cache, callbackURL, showProgress)
	if err != nil {
		return err
	}
	return SimpleConnect(run.GetID(), scraperDirectory, clientServerURL, outputFile, cache, eventCallback, showProgress)
}

func SimpleStart(runID string, scraperDirectory string, clientServerURL string, environment map[string]string,
	outputFile string, cache bool, callbackURL string, showProgress bool) error {
	run := &Run{Client: New(clientServerURL), Run: protocol.Run{ID: runID}}

	// Upload the app
	if showProgress {
		fmt.Println("[Uploading code]")
	}
	if err := run.PutAppFromDirectory(scraperDirectory, []string{cacheName}); err != nil {
		return err
	}
	// Upload the cache
	if cache {
		if showProgress {
			fmt.Println("[Uploading build cache]")
		}
		if err := uploadCacheIfExists(run, filepath.Join(scraperDirectory, cacheName)); err != nil {
			return err
		}
	}
	if showProgress {
		fmt.Println("[Starting run]")
	}
	// Start the run
	return run.Start(&protocol.StartRunOptions{
		Output:   outputFile,
		Callback: protocol.Callback{URL: callbackURL},
		Env:      reformatEnvironmentVariables(environment),
	})
}

// SimpleConnect connects to a run that has been started and handles the rest
func SimpleConnect(runID string, scraperDirectory string, clientServerURL string, outputFile string, cache bool, eventCallback func(event protocol.Event) error, showProgress bool) error {
	run := &Run{Client: New(clientServerURL), Run: protocol.Run{ID: runID}}

	// Listen for events
	events, err := run.GetEvents("")
	if err != nil {
		return err
	}
	for events.More() {
		e, err := events.Next()
		if err != nil {
			return err
		}
		if err = eventCallback(e); err != nil {
			return err
		}
	}
	if showProgress {
		fmt.Println("[Downloading output file]")
	}
	if err = downloadOutput(run, scraperDirectory, outputFile); err != nil {
		return err
	}
	// Get the build cache
	if cache {
		if showProgress {
			fmt.Println("[Downloading build cache]")
		}
		if err = run.GetCacheToFile(filepath.Join(scraperDirectory, cacheName)); err != nil {
			return err
		}
	}
	// Delete the run
	if showProgress {
		fmt.Println("[Deleting run]")
	}
	if err = run.Delete(); err != nil {
		return err
	}
	return nil
}
