package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/openaustralia/morph-ng/pkg/clayclient"
	"github.com/shirou/gopsutil/net"
	"github.com/spf13/cobra"
)

func streamLogs(run clayclient.Run, stage string, streamName string, stream io.ReadCloser, c chan error) {
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		run.CreateLogEvent(clayclient.LogEvent{Stage: stage, Stream: streamName, Text: scanner.Text()})
	}
	c <- scanner.Err()
}

func runExternalCommand(run clayclient.Run, stage string, commandString string) (clayclient.ExitDataStage, error) {
	var exitData clayclient.ExitDataStage

	// Nasty hacky way to split buildCommand up
	commandParts := strings.Split(commandString, " ")
	command := exec.Command(commandParts[0], commandParts[1:]...)
	stdout, err := command.StdoutPipe()
	if err != nil {
		return exitData, err
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		return exitData, err
	}
	// Capture the time and the network counters
	start := time.Now()
	statsStart, err := net.IOCounters(false)
	if err != nil {
		return exitData, err
	}
	// Since we're asking for the aggregates we should only ever receive one answer
	if len(statsStart) != 1 {
		return exitData, errors.New("Only expected one stat")
	}

	if err := command.Start(); err != nil {
		return exitData, err
	}

	c := make(chan error)
	go streamLogs(run, stage, "stdout", stdout, c)
	go streamLogs(run, stage, "stderr", stderr, c)
	err = <-c
	if err != nil {
		return exitData, err
	}
	err = <-c
	if err != nil {
		return exitData, err
	}

	err = command.Wait()
	if err != nil {
		return exitData, err
	}
	statsEnd, err := net.IOCounters(false)
	if err != nil {
		return exitData, err
	}
	// Since we're asking for the aggregates we should only ever receive one answer
	if len(statsEnd) != 1 {
		return exitData, errors.New("Only expected one stat")
	}

	exitData.Usage.NetworkIn = statsEnd[0].BytesRecv - statsStart[0].BytesRecv
	exitData.Usage.NetworkOut = statsEnd[0].BytesSent - statsStart[0].BytesSent
	exitData.Usage.WallTime = time.Now().Sub(start).Seconds()
	exitData.Usage.CPUTime = command.ProcessState.UserTime().Seconds() +
		command.ProcessState.SystemTime().Seconds()
	// This bit will only return something when run on Linux I think
	rusage, ok := command.ProcessState.SysUsage().(*syscall.Rusage)
	if ok {
		exitData.Usage.MaxRSS = rusage.Maxrss
	}
	exitData.ExitCode = command.ProcessState.ExitCode()

	return exitData, nil
}

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
		err := run.CreateStartEvent(clayclient.StartEvent{Stage: "build"})
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
		var exitData clayclient.ExitData

		exitData.Build, err = runExternalCommand(run, "build", buildCommand)
		if err != nil {
			log.Fatal(err)
		}

		// TODO: Check the exit code of the build stage
		// Send the build finished event immediately when the build command has finished
		// Effectively the cache uploading happens between the build and run stages
		err = run.CreateFinishEvent(clayclient.FinishEvent{Stage: "build"})
		if err != nil {
			log.Fatal(err)
		}

		// Temporarily (for the purposes of making the tests easier in the short term)
		// if the cache directory is empty then don't try upload it
		// TODO: Get rid of this check and update the tests
		files, err := ioutil.ReadDir("/tmp/cache")
		if err != nil {
			log.Fatal(err)
		}
		if len(files) > 0 {
			err = run.PutCacheFromDirectory("/tmp/cache")
			if err != nil {
				log.Fatal(err)
			}
		}

		err = run.CreateStartEvent(clayclient.StartEvent{Stage: "run"})
		if err != nil {
			log.Fatal(err)
		}

		exitData.Run, err = runExternalCommand(run, "run", runCommand)
		if err != nil {
			log.Fatal(err)
		}

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

		err = run.CreateFinishEvent(clayclient.FinishEvent{Stage: "run"})
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
