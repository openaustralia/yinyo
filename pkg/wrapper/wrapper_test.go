package wrapper

// This tests the "yinyo wrapper" without running it in a kubernetes cluster

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	mocks "github.com/openaustralia/yinyo/mocks/pkg/apiclient"
	"github.com/openaustralia/yinyo/pkg/protocol"
	"github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createTemporaryDirectories() (appPath string, importPath string, cachePath string, envPath string) {
	appPath, _ = ioutil.TempDir("", "app")
	importPath, _ = ioutil.TempDir("", "import")
	cachePath, _ = ioutil.TempDir("", "cache")
	envPath, _ = ioutil.TempDir("", "env")
	return
}

func TestSimpleRun(t *testing.T) {
	appPath, importPath, cachePath, envPath := createTemporaryDirectories()
	defer os.RemoveAll(appPath)
	defer os.RemoveAll(importPath)
	defer os.RemoveAll(cachePath)
	defer os.RemoveAll(envPath)

	run := new(mocks.RunInterface)
	run.On("CreateStartEvent", "build").Return(nil)
	run.On("GetAppToDirectory", importPath).Return(nil).Run(func(args mock.Arguments) {
		copy.Copy("fixtures/scrapers/hello-world", importPath)
	})
	run.On("GetCacheToDirectory", cachePath).Return(nil).Run(func(args mock.Arguments) {
		copy.Copy("fixtures/scrapers/hello-world", importPath)
	})
	run.On("CreateLogEvent", "build", "stdout", "_app_").Return(nil)
	run.On("CreateLogEvent", "build", "stdout", "Procfile").Return(nil)
	run.On("CreateLogEvent", "build", "stdout", "requirements.txt").Return(nil)
	run.On("CreateLogEvent", "build", "stdout", "runtime.txt").Return(nil)
	run.On("CreateLogEvent", "build", "stdout", "scraper.py").Return(nil)
	run.On("CreateLogEvent", "build", "stdout", "_cache_").Return(nil)
	run.On("CreateLogEvent", "build", "stdout", "requirements.txt").Return(nil)
	run.On("CreateLogEvent", "build", "stdout", "runtime.txt").Return(nil)
	run.On("CreateLogEvent", "build", "stdout", "scraper.py").Return(nil)
	run.On("CreateFinishEvent", "build").Return(nil)
	run.On("PutCacheFromDirectory", cachePath).Return(nil)
	run.On("CreateStartEvent", "run").Return(nil)
	run.On("CreateLogEvent", "run", "stdout", "Ran").Return(nil)
	run.On("PutExitData", mock.MatchedBy(func(e protocol.ExitData) bool {
		// Check that the exit codes are something sensible
		return e.Build.ExitCode == 0 &&
			e.Run.ExitCode == 0 &&
			// The usage values are going to be a little different each time. So, the best we
			// can do for the moment is just check that they are not zero
			// Also not checking network usage because it will be non-zero when run under Linux and zero when run on OS X
			e.Build.Usage.WallTime > 0 &&
			e.Build.Usage.CPUTime > 0 &&
			e.Build.Usage.MaxRSS > 0 &&
			e.Run.Usage.WallTime > 0 &&
			e.Run.Usage.CPUTime > 0 &&
			e.Run.Usage.MaxRSS > 0
	})).Return(nil)
	run.On("PutOutputFromFile", filepath.Join(appPath, "output.txt")).Return(nil)
	run.On("CreateFinishEvent", "run").Return(nil)
	run.On("CreateLastEvent").Return(nil)

	err := Run(run, &Options{
		ImportPath:   importPath,
		CachePath:    cachePath,
		AppPath:      appPath,
		EnvPath:      envPath,
		BuildCommand: `bash -c "echo _app_; ls ` + importPath + `; echo _cache_; ls ` + cachePath + `"`,
		RunCommand:   "echo Ran",
		RunOutput:    "output.txt",
	})
	if err != nil {
		log.Fatal(err)
	}
	run.AssertExpectations(t)
	// TODO: Test that output is correctly uploaded
}

func TestEnvironmentVariables(t *testing.T) {
	appPath, importPath, cachePath, envPath := createTemporaryDirectories()
	defer os.RemoveAll(appPath)
	defer os.RemoveAll(importPath)
	defer os.RemoveAll(cachePath)
	defer os.RemoveAll(envPath)

	run := new(mocks.RunInterface)
	run.On("CreateStartEvent", "build").Return(nil)
	run.On("GetAppToDirectory", importPath).Return(nil)
	run.On("GetCacheToDirectory", cachePath).Return(nil)
	run.On("CreateLogEvent", "build", "stdout", "Build").Return(nil)
	run.On("CreateFinishEvent", "build").Return(nil)
	run.On("PutCacheFromDirectory", cachePath).Return(nil)
	run.On("CreateStartEvent", "run").Return(nil)
	run.On("CreateLogEvent", "run", "stdout", "Run").Return(nil)
	run.On("PutExitData", mock.Anything).Return(nil)
	run.On("CreateFinishEvent", "run").Return(nil)
	run.On("CreateLastEvent").Return(nil)

	err := Run(run, &Options{
		ImportPath:   importPath,
		CachePath:    cachePath,
		AppPath:      appPath,
		EnvPath:      envPath,
		Environment:  map[string]string{"FOO": "bar"},
		BuildCommand: `echo Build`,
		RunCommand:   `echo Run`,
	})
	if err != nil {
		log.Fatal(err)
	}
	// Check that environment files have been set up correctly
	b, err := ioutil.ReadFile(filepath.Join(envPath, "FOO"))
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, "bar", string(b))
	run.AssertExpectations(t)

}

func TestFailingBuild(t *testing.T) {
	appPath, importPath, cachePath, envPath := createTemporaryDirectories()
	defer os.RemoveAll(appPath)
	defer os.RemoveAll(importPath)
	defer os.RemoveAll(cachePath)
	defer os.RemoveAll(envPath)

	run := new(mocks.RunInterface)
	run.On("CreateStartEvent", "build").Return(nil)
	run.On("GetAppToDirectory", importPath).Return(nil).Run(func(args mock.Arguments) {
		copy.Copy("fixtures/scrapers/hello-world", importPath)
	})
	// Let the client know that there is no cache in this case
	run.On("GetCacheToDirectory", cachePath).Return(errors.New("404 Not Found"))
	run.On("CreateLogEvent", "build", "stderr", "bash: failing_command: command not found").Return(nil)
	run.On("CreateFinishEvent", "build").Return(nil)
	run.On("PutCacheFromDirectory", cachePath).Return(nil)
	run.On("PutExitData", mock.MatchedBy(func(e protocol.ExitData) bool {
		// Check that the exit codes are something sensible
		return e.Build.ExitCode == 127 &&
			// The usage values are going to be a little different each time. So, the best we
			// can do for the moment is just check that they are not zero
			// Also not checking network usage because it will be non-zero when run under Linux and zero when run on OS X
			e.Build.Usage.WallTime > 0 &&
			e.Build.Usage.CPUTime > 0 &&
			e.Build.Usage.MaxRSS > 0 &&
			e.Run == nil
	})).Return(nil)
	run.On("CreateLastEvent").Return(nil)

	err := Run(run, &Options{
		ImportPath:   importPath,
		CachePath:    cachePath,
		AppPath:      appPath,
		EnvPath:      envPath,
		BuildCommand: `bash -c "failing_command"`,
		RunCommand:   "echo Ran",
		RunOutput:    "output.txt",
	})
	if err != nil {
		log.Fatal(err)
	}
	run.AssertExpectations(t)
}

func TestFailingRun(t *testing.T) {
	appPath, importPath, cachePath, envPath := createTemporaryDirectories()
	defer os.RemoveAll(appPath)
	defer os.RemoveAll(importPath)
	defer os.RemoveAll(cachePath)
	defer os.RemoveAll(envPath)

	run := new(mocks.RunInterface)
	run.On("CreateStartEvent", "build").Return(nil)
	run.On("GetAppToDirectory", importPath).Return(nil).Run(func(args mock.Arguments) {
		copy.Copy("fixtures/scrapers/hello-world", importPath)
	})
	// Let the client know that there is no cache in this case
	run.On("GetCacheToDirectory", cachePath).Return(errors.New("404 Not Found"))
	run.On("CreateLogEvent", "build", "stdout", "build").Return(nil)
	run.On("CreateFinishEvent", "build").Return(nil)
	run.On("PutCacheFromDirectory", cachePath).Return(nil)
	run.On("CreateStartEvent", "run").Return(nil)
	run.On("CreateLogEvent", "run", "stderr", "bash: failing_command: command not found").Return(nil)
	run.On("PutExitData", mock.MatchedBy(func(e protocol.ExitData) bool {
		// Check that the exit codes are something sensible
		return e.Build.ExitCode == 0 &&
			e.Run.ExitCode == 127 &&
			// The usage values are going to be a little different each time. So, the best we
			// can do for the moment is just check that they are not zero
			// Also not checking network usage because it will be non-zero when run under Linux and zero when run on OS X
			e.Build.Usage.WallTime > 0 &&
			e.Build.Usage.CPUTime > 0 &&
			e.Build.Usage.MaxRSS > 0 &&
			e.Run.Usage.WallTime > 0 &&
			e.Run.Usage.CPUTime > 0 &&
			e.Run.Usage.MaxRSS > 0
	})).Return(nil)
	run.On("PutOutputFromFile", filepath.Join(appPath, "output.txt")).Return(nil)
	run.On("CreateFinishEvent", "run").Return(nil)
	run.On("CreateLastEvent").Return(nil)

	err := Run(run, &Options{
		ImportPath:   importPath,
		CachePath:    cachePath,
		AppPath:      appPath,
		EnvPath:      envPath,
		BuildCommand: `bash -c "echo build"`,
		// Send something to the output file then fail
		RunCommand: `bash -c "cd ` + appPath + `; echo hello > output.txt; failing_command"`,
		RunOutput:  "output.txt",
	})
	if err != nil {
		log.Fatal(err)
	}
	run.AssertExpectations(t)
}

func TestInternalError(t *testing.T) {
	appPath, importPath, cachePath, envPath := createTemporaryDirectories()
	defer os.RemoveAll(appPath)
	defer os.RemoveAll(importPath)
	defer os.RemoveAll(cachePath)
	defer os.RemoveAll(envPath)

	run := new(mocks.RunInterface)
	run.On("CreateStartEvent", "build").Return(nil)
	// Let's simulate an error with the blob storage. So, the wrapper is trying to
	// get the application and there's a problem.
	run.On("GetAppToDirectory", importPath).Return(errors.New("Something went wrong"))
	run.On("CreateLogEvent", "", "interr", "Internal error. The run will be automatically restarted.").Return(nil)

	err := Run(run, &Options{
		ImportPath:   importPath,
		CachePath:    cachePath,
		AppPath:      appPath,
		EnvPath:      envPath,
		BuildCommand: "echo Build",
		RunCommand:   "echo Ran",
		RunOutput:    "output.txt",
	})
	// Because we expect the command to fail
	assert.NotNil(t, err)
	run.AssertExpectations(t)
}
