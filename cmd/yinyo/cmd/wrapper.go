package cmd

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/kballard/go-shellquote"
	"github.com/openaustralia/yinyo/pkg/event"
	"github.com/openaustralia/yinyo/pkg/yinyoclient"
	"github.com/shirou/gopsutil/net"
	"github.com/spf13/cobra"
)

func streamLogs(run yinyoclient.Run, stage string, streamName string, stream io.ReadCloser, c chan error, eventsChan chan event.Event) {
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		eventsChan <- event.NewLogEvent("", stage, streamName, scanner.Text())
	}
	c <- scanner.Err()
}

var wg sync.WaitGroup

func eventsSender(run yinyoclient.Run, eventsChan <-chan event.Event) {
	defer wg.Done()

	// TODO: Send all events in a single http request
	for event := range eventsChan {
		run.CreateEvent(event)
	}
}

// env is an array of strings to set environment variables to in the form "VARIABLE=value", ...
func runExternalCommand(run yinyoclient.Run, stage string, commandString string, env []string) (yinyoclient.ExitDataStage, error) {
	// make a channel with a capacity of 100.
	eventsChan := make(chan event.Event, 1000)

	wg.Add(1)
	// start the worker that sends the event messages
	go eventsSender(run, eventsChan)

	var exitData yinyoclient.ExitDataStage

	// Splits string up into pieces using shell rules
	commandParts, err := shellquote.Split(commandString)
	if err != nil {
		return exitData, err
	}
	command := exec.Command(commandParts[0], commandParts[1:]...)
	// Add the environment variables to the pre-existing environment
	// TODO: Do we want to zero out the environment?
	command.Env = append(os.Environ(), env...)
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
	go streamLogs(run, stage, "stdout", stdout, c, eventsChan)
	go streamLogs(run, stage, "stderr", stderr, c, eventsChan)
	err = <-c
	if err != nil {
		return exitData, err
	}
	err = <-c
	if err != nil {
		return exitData, err
	}

	// Now wait for all the events to get sent via http
	close(eventsChan)
	wg.Wait()

	err = command.Wait()
	if err != nil && command.ProcessState.ExitCode() == 0 {
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

var appPath, importPath, cachePath, runOutput, serverURL, buildCommand, runCommand string
var wrapperEnvironment map[string]string

func init() {
	wrapperCmd.Flags().StringVar(&appPath, "app", "/app", "herokuish app path")
	wrapperCmd.Flags().StringVar(&importPath, "import", "/tmp/app", "herokuish import path")
	wrapperCmd.Flags().StringVar(&cachePath, "cache", "/tmp/cache", "herokuish cache path")
	wrapperCmd.Flags().StringVar(&runOutput, "output", "", "relative path to output file")
	wrapperCmd.Flags().StringVar(&serverURL, "server", "http://yinyo-server.yinyo-system:8080", "override yinyo server URL")
	wrapperCmd.Flags().StringVar(&buildCommand, "buildcommand", "/bin/herokuish buildpack build", "override the herokuish build command (for testing)")
	wrapperCmd.Flags().StringVar(&runCommand, "runcommand", "/bin/herokuish procfile start scraper", "override the herokuish run command (for testing)")
	wrapperCmd.Flags().StringToStringVar(&wrapperEnvironment, "env", map[string]string{}, "Set one or more environment variables (e.g. --env foo=twiddle,bar=blah)")
	rootCmd.AddCommand(wrapperCmd)
}

func checkError(err error, run yinyoclient.Run, stage string, text string) {
	if err != nil {
		run.CreateEvent(event.NewLogEvent("", "build", "interr", "There was a problem getting the code"))
		log.Fatal(err)
	}
}

var wrapperCmd = &cobra.Command{
	Use:   "wrapper run_name run_token",
	Short: "Manages the building and running of a scraper",
	Long:  "Manages the building and running of a scraper inside a container. Used internally by the system.",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runName := args[0]
		runToken := args[1]

		client := yinyoclient.New(serverURL)
		run := yinyoclient.Run{Name: runName, Token: runToken, Client: client}
		err := run.CreateEvent(event.NewStartEvent("", "build"))
		checkError(err, run, "build", "Could not create event")

		// Create and populate herokuish import path and cache path
		err = os.MkdirAll(importPath, 0755)
		checkError(err, run, "build", "Could not create directory")
		err = os.MkdirAll(cachePath, 0755)
		checkError(err, run, "build", "Could not create directory")
		// TODO: Allow envPath to be changed by a flag
		envPath := "/tmp/env"
		err = os.MkdirAll(envPath, 0755)
		checkError(err, run, "build", "Could not create directory")

		// Write the environment variables to /tmp/env in the format defined by the buildpack API
		for name, value := range wrapperEnvironment {
			f, err := os.Create(filepath.Join(envPath, name))
			checkError(err, run, "build", "Could not create environment file")
			_, err = f.WriteString(value)
			checkError(err, run, "build", "Could not write to environment file")
			f.Close()
		}

		err = run.GetAppToDirectory(importPath)
		checkError(err, run, "build", "Could not get the code")
		d1 := []byte("scraper: /bin/start.sh")
		err = ioutil.WriteFile(filepath.Join(importPath, "Procfile"), d1, 0644)
		checkError(err, run, "build", "Could not write to a file")
		// If the cache doesn't exit this will not error
		err = run.GetCacheToDirectory(cachePath)
		checkError(err, run, "build", "Could not get the cache")

		env := []string{
			"APP_PATH=" + appPath,
			"CACHE_PATH=" + cachePath,
			"IMPORT_PATH=" + importPath,
		}

		// Initially do a very naive way of calling the command just to get things going
		var exitData yinyoclient.ExitData

		exitDataStage, err := runExternalCommand(run, "build", buildCommand, env)
		checkError(err, run, "build", "Unexpected error while building")
		exitData.Build = &exitDataStage

		// Send the build finished event immediately when the build command has finished
		// Effectively the cache uploading happens between the build and run stages
		err = run.CreateEvent(event.NewFinishEvent("", "build"))
		checkError(err, run, "build", "Could not create event")

		err = run.PutCacheFromDirectory(cachePath)
		// TODO: We're not actually in the "build" stage here
		checkError(err, run, "build", "Could not upload cache")

		// Only do the main run if the build was succesful
		if exitData.Build.ExitCode == 0 {
			err = run.CreateEvent(event.NewStartEvent("", "run"))
			checkError(err, run, "run", "Could not create event")

			exitDataStage, err := runExternalCommand(run, "run", runCommand, env)
			checkError(err, run, "run", "Unexpected error while running")
			exitData.Run = &exitDataStage

			err = run.PutExitData(exitData)
			checkError(err, run, "run", "Could not upload exit data")

			if runOutput != "" {
				err = run.PutOutputFromFile(filepath.Join(appPath, runOutput))
				checkError(err, run, "run", "Could not upload output")
			}

			err = run.CreateEvent(event.NewFinishEvent("", "run"))
			checkError(err, run, "run", "Could not create event")
		} else {
			// TODO: Only upload the exit data for the build
			err := run.PutExitData(exitData)
			checkError(err, run, "run", "Could not upload exit data")
		}

		err = run.CreateEvent(event.NewLastEvent(""))
		checkError(err, run, "run", "Could not create event")
	},
}
