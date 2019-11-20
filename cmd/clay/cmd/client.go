package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/openaustralia/morph-ng/pkg/clayclient"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(clientCmd)
}

var clientCmd = &cobra.Command{
	Use:   "client scraper_directory output_file [callback url]",
	Short: "Runs a scraper in a local directory using clay",
	Long: `Runs a scraper in a local directory using clay
The output is written to the same local directory at the end. The output file path
is given relative to the scraper directory
Optionally provide a callback url. For every event on the scraper this will get called.
Note: To be able to authenticate the callback you'll need to specify a secret in the url.
Something like http://my-url-endpoint.com?key=special-secret-stuff would do the trick`,
	Run: func(cmd *cobra.Command, args []string) {
		scraperDirectory := args[0]
		// TODO: Make the output file optional
		// TODO: Check that the output file is a relative path and if not error
		outputFile := args[1]
		// callbackURL := args[2]

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
		// TODO: Chop off any trailing "/" in scraperDirectory to make cachePath
		// TODO: Use filepath.Join here instead. Will solve above problem
		cachePath := fmt.Sprintf("assets/client-storage/cache/%v.tgz", scraperDirectory)
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
		// TODO: Add callback
		err = run.Start(&clayclient.StartRunOptions{Output: outputFile})
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
		// TODO: Extract this into a method GetOutputToFile
		output, err := run.GetOutput()
		if err != nil {
			log.Fatal(err)
		}

		f, err := os.Create(filepath.Join(scraperDirectory, outputFile))
		if err != nil {
			log.Fatal(err)
		}
		io.Copy(f, output)
		f.Close()
		output.Close()

		// Get the build cache
		// TODO: Extract to GetCacheToFile
		cache, err := run.GetCache()
		if err != nil {
			log.Fatal(err)
		}
		// Create the directory to store the cache if it doesn't already exist
		// TODO: This actually creates one directory too many?
		err = os.MkdirAll(filepath.Join("assets/client-storage/cache", scraperDirectory), 0755)
		if err != nil {
			log.Fatal(err)
		}
		f, err = os.Create(cachePath)
		if err != nil {
			log.Fatal(err)
		}
		io.Copy(f, cache)

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
