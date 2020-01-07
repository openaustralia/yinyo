package apiclient

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/openaustralia/yinyo/pkg/protocol"
)

const cacheName = ".yinyo-build-cache.tgz"

func uploadCacheIfExists(run Run, cachePath string) error {
	file, err := os.Open(cachePath)
	if err != nil {
		// If the cache doesn't exist then skip the uploading bit
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		err = run.PutCache(file)
		if err != nil {
			return err
		}
		file.Close()
	}
	return nil
}

func downloadOutput(run Run, scraperDirectory string, outputFile string) error {
	// Get the run output
	if outputFile != "" {
		err := run.GetOutputToFile(filepath.Join(scraperDirectory, outputFile))
		if err != nil {
			if IsNotFound(err) {
				log.Printf("Warning: output file %v does not exist", outputFile)
			} else {
				return err
			}
		}
	}
	return nil
}

// Simple is a super simple high level way of running a scraper that exists on the local file system
// and displaying the results to stdout/stderr. This is used by the command line client
// It makes a simple common use case a little simpler to implement
func Simple(scraperDirectory string, clientServerURL string, environment map[string]string, outputFile string, callbackURL string, showEventsJSON bool) error {
	client := New(clientServerURL)
	// Create the run
	run, err := client.CreateRun(scraperDirectory)
	if err != nil {
		return err
	}
	// Upload the app
	err = run.PutAppFromDirectory(scraperDirectory, []string{cacheName})
	if err != nil {
		return err
	}
	// Upload the cache
	cachePath := filepath.Join(scraperDirectory, cacheName)
	err = uploadCacheIfExists(run, cachePath)
	if err != nil {
		return err
	}
	var envVariables []protocol.EnvVariable
	for k, v := range environment {
		// TODO: Fix this inefficient way
		envVariables = append(envVariables, protocol.EnvVariable{Name: k, Value: v})
	}
	// Start the run
	err = run.Start(&protocol.StartRunOptions{
		Output:   outputFile,
		Callback: protocol.Callback{URL: callbackURL},
		Env:      envVariables,
	})
	if err != nil {
		return err
	}
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
		err = displayEvent(e, showEventsJSON)
		if err != nil {
			return err
		}
	}
	err = downloadOutput(run, scraperDirectory, outputFile)
	if err != nil {
		return err
	}
	// Get the build cache
	err = run.GetCacheToFile(cachePath)
	if err != nil {
		return err
	}
	// Delete the run
	err = run.Delete()
	if err != nil {
		return err
	}
	return nil
}

func displayEvent(e protocol.Event, showEventsJSON bool) error {
	if showEventsJSON {
		// Convert the event back to JSON for display
		b, err := json.Marshal(e)
		if err != nil {
			return err
		}
		fmt.Println(string(b))
	} else {
		// Only display the log events to the user
		l, ok := e.Data.(protocol.LogData)
		if ok {
			f, err := osStream(l.Stream)
			if err != nil {
				return err
			}
			fmt.Fprintln(f, l.Text)
		}
	}
	return nil
}

// Convert the internal text representation of a stream type ("stdout"/"stderr") to the go stream
// we can write to
func osStream(stream string) (*os.File, error) {
	switch stream {
	// TODO: Extract string constant
	case "stdout":
		return os.Stdout, nil
	case "stderr", "interr":
		return os.Stderr, nil
	default:
		return nil, fmt.Errorf("Unexpected stream %v", stream)
	}
}
