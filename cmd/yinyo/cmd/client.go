package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/openaustralia/yinyo/pkg/apiclient"
	"github.com/openaustralia/yinyo/pkg/protocol"
	"github.com/spf13/cobra"
)

var callbackURL, outputFile, clientServerURL string
var showEventsJSON bool
var environment map[string]string

func init() {
	clientCmd.Flags().StringVar(&callbackURL, "callback", "", "Optionally provide a callback URL. For every event a POST to the URL will be made. To be able to authenticate the callback you'll need to specify a secret in the URL. Something like http://my-url-endpoint.com?key=special-secret-stuff would do the trick")
	// TODO: Check that the output file is a relative path and if not error
	clientCmd.Flags().StringVar(&outputFile, "output", "", "The output is written to the same local directory at the end. The output file path is given relative to the scraper directory")
	clientCmd.Flags().StringVar(&clientServerURL, "server", "http://localhost:8080", "Override yinyo server URL")
	clientCmd.Flags().StringToStringVar(&environment, "env", map[string]string{}, "Set one or more environment variables (e.g. --env foo=twiddle,bar=blah)")
	clientCmd.Flags().BoolVar(&showEventsJSON, "eventsjson", false, "Show the full events output as JSON instead of the default of just showing the log events as text")
	rootCmd.AddCommand(clientCmd)
}

var clientCmd = &cobra.Command{
	Use:   "client scraper_directory",
	Short: "Runs a scraper in a local directory using yinyo",
	Long:  "Runs a scraper in a local directory using yinyo",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scraperDirectory := args[0]
		err := apiclient.Simple(
			scraperDirectory, clientServerURL, environment, outputFile, callbackURL,
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

		if err != nil {
			log.Fatal(err)
		}
	},
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
