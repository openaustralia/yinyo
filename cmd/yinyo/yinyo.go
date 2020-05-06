package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/openaustralia/yinyo/pkg/apiclient"
	"github.com/openaustralia/yinyo/pkg/protocol"
	"github.com/spf13/cobra"
)

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
		return nil, fmt.Errorf("unexpected stream %v", stream)
	}
}

func display(event protocol.Event, showEventsJSON bool) error {
	if showEventsJSON {
		// Convert the event back to JSON for display
		b, err := json.Marshal(event)
		if err != nil {
			return err
		}
		fmt.Println(string(b))
	} else {
		// Only display the log events to the user
		l, ok := event.Data.(protocol.LogData)
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

func apiKeysPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".yinyo"), nil
}

// Write to user's local directory as .yinyo
// Store a separate api key for each server URL
func saveAPIKeys(apiKeys map[string]string) error {
	path, err := apiKeysPath()
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(apiKeys)
}

func loadAPIKeys() (map[string]string, error) {
	apiKeys := make(map[string]string)
	path, err := apiKeysPath()
	if err != nil {
		return apiKeys, err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return apiKeys, nil
		}
		return apiKeys, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	err = dec.Decode(&apiKeys)
	return apiKeys, err
}

func askForAndSaveAPIKey(clientServerURL string) (string, error) {
	fmt.Print("Enter your api key: ")
	var apiKey string
	fmt.Scanln(&apiKey)

	apiKeys, err := loadAPIKeys()
	if err != nil {
		return apiKey, err
	}
	apiKeys[clientServerURL] = apiKey
	err = saveAPIKeys(apiKeys)
	return apiKey, err
}

func run(scraperDirectory string, clientServerURL string, environment map[string]string,
	outputFile string, cache bool, callbackURL string, showEventsJSON bool) {
	apiKeys, err := loadAPIKeys()
	if err != nil {
		log.Fatal(err)
	}
	apiKey := apiKeys[clientServerURL]
	for {
		err := apiclient.Simple(
			scraperDirectory, clientServerURL, environment, outputFile, cache, callbackURL, apiKey,
			func(event protocol.Event) error { return display(event, showEventsJSON) },
		)
		if err != nil {
			// If get unauthorized error back then we should let the user enter their api key and try again
			// And we should save away the api key (for this particular server) for later use
			if apiclient.IsUnauthorized(err) {
				apiKey, err = askForAndSaveAPIKey(clientServerURL)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				log.Fatal(err)
			}
		} else {
			break
		}
	}
}

// Reconnect to existing run
func reconnect(runID string, scraperDirectory string, clientServerURL string, outputFile string, cache bool, showEventsJSON bool) {
	err := apiclient.SimpleConnect(runID, scraperDirectory, clientServerURL, outputFile, cache, func(event protocol.Event) error { return display(event, showEventsJSON) })
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Show the source of the error with the standard logger. Don't show date & time
	log.SetFlags(log.Lshortfile)

	var callbackURL, outputFile, clientServerURL, connectToRunID string
	var showEventsJSON, cache bool
	var environment map[string]string

	var rootCmd = &cobra.Command{
		Use:   "yinyo scraper_directory",
		Short: "Yinyo runs heaps of scrapers easily, fast and scalably",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			scraperDirectory := args[0]

			if connectToRunID == "" {
				run(scraperDirectory, clientServerURL, environment, outputFile, cache, callbackURL, showEventsJSON)
			} else {
				reconnect(connectToRunID, scraperDirectory, clientServerURL, outputFile, cache, showEventsJSON)
			}
		},
	}

	rootCmd.Flags().StringVar(&callbackURL, "callback", "", "Optionally provide a callback URL. For every event a POST to the URL will be made. To be able to authenticate the callback you'll need to specify a secret in the URL. Something like http://my-url-endpoint.com?key=special-secret-stuff would do the trick")
	// TODO: Check that the output file is a relative path and if not error
	rootCmd.Flags().StringVar(&outputFile, "output", "", "The output is written to the same local directory at the end. The output file path is given relative to the scraper directory")
	rootCmd.Flags().StringVar(&connectToRunID, "connect", "", "Connect to a run that has already started by giving the run ID")
	rootCmd.Flags().StringVar(&clientServerURL, "server", "http://localhost:8080", "Override yinyo server URL")
	rootCmd.Flags().StringToStringVar(&environment, "env", map[string]string{}, "Set one or more environment variables (e.g. --env foo=twiddle,bar=blah)")
	rootCmd.Flags().BoolVar(&showEventsJSON, "allevents", false, "Show the full events output as JSON instead of the default of just showing the log events as text")
	rootCmd.Flags().BoolVar(&cache, "cache", false, "Enable the download and upload of the build cache")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
