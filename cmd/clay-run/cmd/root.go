package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/openaustralia/morph-ng/pkg/clayclient"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "clay-run run_name run_output",
	Short: "Builds and runs a scraper",
	Long:  "Builds and runs a scraper and talks back to the Clay server.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runName := args[0]
		runToken := os.Getenv("CLAY_INTERNAL_RUN_TOKEN")
		runOutput := args[1]

		// We allow some settings to be overridden for the purposes of testing.
		// We don't allow users to change any environment variables that start
		// with CLAY_INTERNAL_SERVER_. So they can't change any of these

		// TODO: Convert these special environment variables to command line options
		serverURL, ok := os.LookupEnv("CLAY_INTERNAL_SERVER_URL")
		if !ok {
			serverURL = "http://clay-server.clay-system:8080"
		}

		buildCommand, ok := os.LookupEnv("CLAY_INTERNAL_BUILD_COMMAND")
		if !ok {
			buildCommand = "/bin/herokuish buildpack build"
		}

		runCommand, ok := os.LookupEnv("CLAY_INTERNAL_RUN_COMMAND")
		if !ok {
			runCommand = "/bin/herokuish procfile start scraper"
		}

		fmt.Println("runName", runName)
		fmt.Println("runToken", runToken)
		fmt.Println("runOutput", runOutput)
		fmt.Println("serverURL", serverURL)
		fmt.Println("buildCommand", buildCommand)
		fmt.Println("runCommand", runCommand)

		client := clayclient.New(serverURL)
		run := clayclient.Run{Name: runName, Token: runToken, Client: client}
		err := run.CreateStartEvent("build")
		if err != nil {
			log.Fatal(err)
		}

		// Create and populate /tmp/app and /tmp/cache
		err = os.MkdirAll("/tmp/app", 0755)
		if err != nil {
			log.Fatal(err)
		}
		err = os.MkdirAll("/tmp/cache", 0755)
		if err != nil {
			log.Fatal(err)
		}
		err = run.GetAppToDirectory("/tmp/app")
		if err != nil {
			log.Fatal(err)
		}
		d1 := []byte("scraper: /bin/start.sh")
		err = ioutil.WriteFile("/tmp/app/Procfile", d1, 0644)
		if err != nil {
			log.Fatal(err)
		}
		// TODO: Don't fail if the cache doesn't yet exist
		err = run.GetCacheToDirectory("/tmp/cache")
		if err != nil {
			log.Fatal(err)
		}

		// Initially do a very naive way of calling the command just to get things going
		// TODO: Capture stdout and stderr
		// TODO: Gather usage stats
		// Nasty hacky way to split buildCommand up
		commandParts := strings.Split(buildCommand, " ")
		command := exec.Command(commandParts[0], commandParts[1:]...)
		stdout, err := command.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		if err := command.Start(); err != nil {
			log.Fatal(err)
		}
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			run.CreateLogEvent("build", "stdout", scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		err = run.PutCacheFromDirectory("/tmp/cache")
		if err != nil {
			log.Fatal(err)
		}

		err = run.CreateFinishEvent("build")
		if err != nil {
			log.Fatal(err)
		}

		err = run.CreateStartEvent("run")
		if err != nil {
			log.Fatal(err)
		}

		commandParts = strings.Split(runCommand, " ")
		command = exec.Command(commandParts[0], commandParts[1:]...)
		stdout, err = command.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		if err := command.Start(); err != nil {
			log.Fatal(err)
		}
		scanner = bufio.NewScanner(stdout)
		for scanner.Scan() {
			run.CreateLogEvent("run", "stdout", scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		var exitData clayclient.ExitData
		// TODO: Populate exitData with actual data!
		if err := run.PutExitData(exitData); err != nil {
			log.Fatal(err)
		}

		if _, err := os.Stat(filepath.Join("/app", runOutput)); !os.IsNotExist(err) {
			f, err := os.Open(filepath.Join("/app", runOutput))
			defer f.Close()
			if err != nil {
				log.Fatal(err)
			}
			err = run.PutOutput(f)
			if err != nil {
				log.Fatal(err)
			}
		}

		err = run.CreateFinishEvent("run")
		if err != nil {
			log.Fatal(err)
		}

		err = run.CreateLastEvent()
		if err != nil {
			log.Fatal(err)
		}
	},
}

// Execute makes it all happen
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
