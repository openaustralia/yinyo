package cmd

import (
	"github.com/openaustralia/yinyo/pkg/apiclient"
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
		apiclient.Simple(scraperDirectory, clientServerURL, environment, outputFile, callbackURL, showEventsJSON)
	},
}
