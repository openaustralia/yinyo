package wrapper

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
	"github.com/openaustralia/yinyo/pkg/apiclient"
	"github.com/openaustralia/yinyo/pkg/protocol"
	"github.com/shirou/gopsutil/net"
)

var wg sync.WaitGroup

func eventsSender(run apiclient.RunInterface, eventsChan <-chan protocol.LogData) {
	defer wg.Done()

	// TODO: Send all events in a single http request
	for e := range eventsChan {
		err := run.CreateLogEvent(e.Stage, e.Stream, e.Text)
		if err != nil {
			// If we can't send an event there's not much point in trying to do anything
			// else but log an error locally
			log.Println("Couldn't send event")
		}
	}
}

func streamLogs(run apiclient.RunInterface, stage string, streamName string, stream io.ReadCloser, c chan error, eventsChan chan protocol.LogData) {
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		eventsChan <- protocol.LogData{Stage: stage, Stream: streamName, Text: scanner.Text()}
	}
	c <- scanner.Err()
}

func runExternalCommand(run apiclient.RunInterface, stage string, commandString string, env []string) (*os.ProcessState, error) {
	// make a channel with a capacity of 100.
	eventsChan := make(chan protocol.LogData, 1000)

	wg.Add(1)
	// start the worker that sends the event messages
	go eventsSender(run, eventsChan)

	// Splits string up into pieces using shell rules
	commandParts, err := shellquote.Split(commandString)
	if err != nil {
		return nil, err
	}
	command := exec.Command(commandParts[0], commandParts[1:]...)
	// Add the environment variables to the pre-existing environment
	// TODO: Do we want to zero out the environment?
	command.Env = append(os.Environ(), env...)
	stdout, err := command.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := command.Start(); err != nil {
		return nil, err
	}

	c := make(chan error)
	go streamLogs(run, stage, "stdout", stdout, c, eventsChan)
	go streamLogs(run, stage, "stderr", stderr, c, eventsChan)
	err = <-c
	if err != nil {
		return nil, err
	}
	err = <-c
	if err != nil {
		return nil, err
	}

	// Now wait for all the events to get sent via http
	close(eventsChan)
	wg.Wait()

	err = command.Wait()
	if err != nil && command.ProcessState.ExitCode() == 0 {
		return nil, err
	}
	return command.ProcessState, nil
}

func aggregateCounters() (net.IOCountersStat, error) {
	stats, err := net.IOCounters(false)
	if err != nil {
		return net.IOCountersStat{}, err
	}
	// Since we're asking for the aggregates we should only ever receive one answer
	if len(stats) != 1 {
		return net.IOCountersStat{}, errors.New("only expected one stat")
	}
	return stats[0], nil
}

// env is an array of strings to set environment variables to in the form "VARIABLE=value", ...
func runExternalCommandWithStats(run apiclient.RunInterface, stage string, commandString string, env []string) (protocol.ExitDataStage, error) {
	var exitData protocol.ExitDataStage

	// Capture the time and the network counters
	start := time.Now()
	statsStart, err := aggregateCounters()
	if err != nil {
		return exitData, err
	}

	state, err := runExternalCommand(run, stage, commandString, env)
	if err != nil {
		return exitData, err
	}

	statsEnd, err := aggregateCounters()
	if err != nil {
		return exitData, err
	}

	exitData.Usage.NetworkIn = statsEnd.BytesRecv - statsStart.BytesRecv
	exitData.Usage.NetworkOut = statsEnd.BytesSent - statsStart.BytesSent
	exitData.Usage.WallTime = time.Since(start).Seconds()
	exitData.Usage.CPUTime = state.UserTime().Seconds() + state.SystemTime().Seconds()
	// This bit will only return something when run on Linux I think
	rusage, ok := state.SysUsage().(*syscall.Rusage)
	if ok {
		// rusage.Maxrss is in kilobytes, while exitData.Usage.MaxRSS is in bytes
		exitData.Usage.MaxRSS = uint64(rusage.Maxrss) * 1024
	}
	exitData.ExitCode = state.ExitCode()

	return exitData, nil
}

// Options are parameters required for calling Run
type Options struct {
	ImportPath   string
	CachePath    string
	AppPath      string
	EnvPath      string
	Environment  map[string]string
	BuildCommand string
	RunCommand   string
	RunOutput    string
}

func setup(run apiclient.RunInterface, options Options) error {
	// Create and populate herokuish import path and cache path
	err := os.MkdirAll(options.ImportPath, 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(options.CachePath, 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(options.EnvPath, 0700)
	if err != nil {
		return err
	}

	// Write the environment variables to /tmp/env in the format defined by the buildpack API
	for name, value := range options.Environment {
		f, err := os.Create(filepath.Join(options.EnvPath, name))
		if err != nil {
			return err
		}
		_, err = f.WriteString(value)
		if err != nil {
			return err
		}
		f.Close()
	}

	err = run.GetAppToDirectory(options.ImportPath)
	if err != nil {
		return err
	}
	d1 := []byte("scraper: /bin/start.sh")
	err = ioutil.WriteFile(filepath.Join(options.ImportPath, "Procfile"), d1, 0644)
	if err != nil {
		return err
	}
	// If the cache doesn't exit this will not error
	err = run.GetCacheToDirectory(options.CachePath)
	// It's not an error if cache doesn't exist
	if err != nil && !apiclient.IsNotFound(err) {
		return err
	}
	return nil
}

func runStage(run apiclient.RunInterface, options Options, env []string, exitDataBuild protocol.ExitDataStage) error {
	err := run.CreateStartEvent("run")
	if err != nil {
		return err
	}

	var exitData protocol.ExitData
	exitData.Build = &exitDataBuild
	exitDataStage, err := runExternalCommandWithStats(run, "run", options.RunCommand, env)
	if err != nil {
		return err
	}
	exitData.Run = &exitDataStage

	err = run.PutExitData(exitData)
	if err != nil {
		return err
	}

	if options.RunOutput != "" {
		err = run.PutOutputFromFile(filepath.Join(options.AppPath, options.RunOutput))
		if err != nil {
			return err
		}
	}

	err = run.CreateFinishEvent("run")
	if err != nil {
		return err
	}
	return nil
}

func runWithError(run apiclient.RunInterface, options Options) error {
	err := run.CreateStartEvent("build")
	if err != nil {
		return err
	}

	err = setup(run, options)
	if err != nil {
		return err
	}

	env := []string{
		"APP_PATH=" + options.AppPath,
		"CACHE_PATH=" + options.CachePath,
		"IMPORT_PATH=" + options.ImportPath,
	}

	// Initially do a very naive way of calling the command just to get things going
	var exitData protocol.ExitData

	exitDataStage, err := runExternalCommandWithStats(run, "build", options.BuildCommand, env)
	if err != nil {
		return err
	}
	exitData.Build = &exitDataStage

	// Send the build finished event immediately when the build command has finished
	// Effectively the cache uploading happens between the build and run stages
	err = run.CreateFinishEvent("build")
	if err != nil {
		return err
	}

	err = run.PutCacheFromDirectory(options.CachePath)
	if err != nil {
		return err
	}

	// Only do the main run if the build was successful
	if exitData.Build.ExitCode == 0 {
		err := runStage(run, options, env, *exitData.Build)
		if err != nil {
			return err
		}
	} else {
		// TODO: Only upload the exit data for the build
		err := run.PutExitData(exitData)
		if err != nil {
			return err
		}
	}

	err = run.CreateLastEvent()
	if err != nil {
		return err
	}
	return nil
}

// Run runs a scraper from inside a container
func Run(run apiclient.RunInterface, options Options) error {
	err := runWithError(run, options)
	if err != nil {
		// Notice that for an internal error we're not logging the stage. We leave that empty.
		//nolint:errcheck // ignore errors while logging error
		//skipcq: GSC-G104
		run.CreateLogEvent("", "interr", "Internal error. The run will be automatically restarted.")
	}
	return err
}
