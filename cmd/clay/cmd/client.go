package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/openaustralia/morph-ng/pkg/clayclient"
	"github.com/spf13/cobra"
)

var callbackURL, outputFile string

func init() {
	clientCmd.Flags().StringVar(&callbackURL, "callback", "", "Optionally provide a callback URL. For every event a POST to the URL will be made. To be able to authenticate the callback you'll need to specify a secret in the URL. Something like http://my-url-endpoint.com?key=special-secret-stuff would do the trick")
	// TODO: Check that the output file is a relative path and if not error
	clientCmd.Flags().StringVar(&outputFile, "output", "", "The output is written to the same local directory at the end. The output file path is given relative to the scraper directory")
	rootCmd.AddCommand(clientCmd)
}

var clientCmd = &cobra.Command{
	Use:   "client scraper_directory",
	Short: "Runs a scraper in a local directory using clay",
	Long:  "Runs a scraper in a local directory using clay",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scraperDirectory := args[0]

		client := clayclient.New("http://localhost:8080")
		// Create the run
		run, err := client.CreateRun(scraperDirectory)
		if err != nil {
			log.Fatal(err)
		}

		// Upload the app
		err = run.PutAppFromDirectory(scraperDirectory)
		if err != nil {
			log.Fatal(err)
		}

		// Upload the cache
		cachePath := filepath.Join("assets/client-storage/cache", scraperDirectory) + ".tgz"
		file, err := os.Open(cachePath)
		if err != nil {
			// If the cache doesn't exist then skip the uploading bit
			if !os.IsNotExist(err) {
				log.Fatal(err)
			}
		} else {
			err = run.PutCache(file)
			if err != nil {
				log.Fatal(err)
			}
			file.Close()
		}

		// Start the run
		// TODO: Add support for setting environment variables
		err = run.Start(&clayclient.StartRunOptions{
			Output:   outputFile,
			Callback: clayclient.Callback{URL: callbackURL},
		})
		if err != nil {
			log.Fatal(err)
		}

		// Listen for events
		events, err := run.GetEvents()
		if err != nil {
			log.Fatal(err)
		}
		for events.More() {
			event, err := events.Next()
			if err != nil {
				log.Fatal(err)
			}
			// Only display the log events to the user
			l, ok := event.(clayclient.LogEvent)
			if ok {
				f, err := osStream(l.Stream)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Fprintln(f, l.Text)
			}
		}

		// Get the run output
		if outputFile != "" {
			path := filepath.Join(scraperDirectory, outputFile)
			err = run.GetOutputToFile(path)
			if err != nil {
				if clayclient.IsNotFound(err) {
					log.Printf("Warning: output file %v does not exist", outputFile)
				} else {
					log.Fatal(err)
				}
			}
		}

		// Get the build cache
		// Create the directory to store the cache if it doesn't already exist
		// TODO: This actually creates one directory too many?
		err = os.MkdirAll(filepath.Join("assets/client-storage/cache", scraperDirectory), 0755)
		err = run.GetCacheToFile(cachePath)
		if err != nil {
			log.Fatal(err)
		}

		// Get the exit data
		// exitData, err := run.GetExitData()
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// fmt.Printf("%+v", exitData)

		// Delete the run
		err = run.Delete()
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
	case "stderr":
		return os.Stderr, nil
	default:
		return nil, fmt.Errorf("Unexpected stream %v", stream)
	}
}
