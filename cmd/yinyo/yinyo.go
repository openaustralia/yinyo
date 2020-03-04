package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

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

func run(scraperDirectory string, clientServerURL string, environment map[string]string, outputFile string, callbackURL string, showEventsJSON bool, apiKey string) error {
	return apiclient.Simple(
		scraperDirectory, clientServerURL, environment, outputFile, callbackURL, apiKey,
		func(event protocol.Event) error {
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
		},
	)

}

func main() {
	// Show the source of the error with the standard logger. Don't show date & time
	log.SetFlags(log.Lshortfile)

	var callbackURL, outputFile, clientServerURL string
	var showEventsJSON bool
	var environment map[string]string

	var rootCmd = &cobra.Command{
		Use:   "yinyo scraper_directory",
		Short: "Yinyo runs heaps of scrapers easily, fast and scalably",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			scraperDirectory := args[0]
			// TODO: Get this value from a file in the user's home directory
			apiKey := ""
			for {
				err := run(scraperDirectory, clientServerURL, environment, outputFile, callbackURL, showEventsJSON, apiKey)
				if err != nil {
					// If get unauthorized error back then we should let the user enter their api key and try again
					// And we should save away the api key (for this particular server) for later use
					if apiclient.IsUnauthorized(err) {
						fmt.Print("Enter your api key: ")
						fmt.Scanln(&apiKey)
					} else {
						log.Fatal(err)
					}
				} else {
					break
				}

			}
		},
	}

	rootCmd.Flags().StringVar(&callbackURL, "callback", "", "Optionally provide a callback URL. For every event a POST to the URL will be made. To be able to authenticate the callback you'll need to specify a secret in the URL. Something like http://my-url-endpoint.com?key=special-secret-stuff would do the trick")
	// TODO: Check that the output file is a relative path and if not error
	rootCmd.Flags().StringVar(&outputFile, "output", "", "The output is written to the same local directory at the end. The output file path is given relative to the scraper directory")
	rootCmd.Flags().StringVar(&clientServerURL, "server", "http://localhost:8080", "Override yinyo server URL")
	rootCmd.Flags().StringToStringVar(&environment, "env", map[string]string{}, "Set one or more environment variables (e.g. --env foo=twiddle,bar=blah)")
	rootCmd.Flags().BoolVar(&showEventsJSON, "eventsjson", false, "Show the full events output as JSON instead of the default of just showing the log events as text")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
