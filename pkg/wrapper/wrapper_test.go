package wrapper

// This tests the "yinyo wrapper" without running it in a kubernetes cluster

import (
	"errors"
	"io/ioutil"
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
	run.On("CreateStartEvent", "build").Return(10, nil)
	run.On("GetAppToDirectory", importPath).Return(nil).Run(func(args mock.Arguments) {
		copy.Copy("fixtures/scrapers/hello-world", importPath)
	})
	run.On("GetCacheToDirectory", cachePath).Return(nil).Run(func(args mock.Arguments) {
		copy.Copy("fixtures/scrapers/hello-world", importPath)
	})
	run.On("CreateLogEvent", "build", "stdout", "_app_").Return(10, nil)
	run.On("CreateLogEvent", "build", "stdout", "Procfile").Return(10, nil)
	run.On("CreateLogEvent", "build", "stdout", "requirements.txt").Return(10, nil)
	run.On("CreateLogEvent", "build", "stdout", "runtime.txt").Return(10, nil)
	run.On("CreateLogEvent", "build", "stdout", "scraper.py").Return(10, nil)
	run.On("CreateLogEvent", "build", "stdout", "_cache_").Return(10, nil)
	run.On("CreateLogEvent", "build", "stdout", "requirements.txt").Return(10, nil)
	run.On("CreateLogEvent", "build", "stdout", "runtime.txt").Return(10, nil)
	run.On("CreateLogEvent", "build", "stdout", "scraper.py").Return(10, nil)
	run.On("CreateFinishEvent", "build", mock.MatchedBy(func(e protocol.ExitDataStage) bool {
		return e.ExitCode == 0 &&
			// The usage values are going to be a little different each time. So, the best we
			// can do for the moment is just check that they are not zero
			// Also not checking network usage because it will be non-zero when run under Linux and zero when run on OS X
			e.Usage.WallTime > 0 &&
			e.Usage.CPUTime > 0 &&
			e.Usage.MaxRSS > 0
	})).Return(10, nil)
	run.On("PutCacheFromDirectory", cachePath).Return(nil)
	run.On("CreateStartEvent", "run").Return(10, nil)
	run.On("CreateLogEvent", "run", "stdout", "Ran").Return(10, nil)
	run.On("PutOutputFromFile", filepath.Join(appPath, "output.txt")).Return(nil)
	run.On("CreateFinishEvent", "run", mock.MatchedBy(func(e protocol.ExitDataStage) bool {
		// Check that the exit codes are something sensible
		return e.ExitCode == 0 &&
			// The usage values are going to be a little different each time. So, the best we
			// can do for the moment is just check that they are not zero
			// Also not checking network usage because it will be non-zero when run under Linux and zero when run on OS X
			e.Usage.WallTime > 0 &&
			e.Usage.CPUTime > 0 &&
			e.Usage.MaxRSS > 0

	})).Return(10, nil)
	run.On("CreateLastEvent").Return(10, nil)

	err := Run(run, &Options{
		ImportPath:   importPath,
		CachePath:    cachePath,
		AppPath:      appPath,
		EnvPath:      envPath,
		BuildCommand: `bash -c "echo _app_; ls ` + importPath + `; echo _cache_; ls ` + cachePath + `"`,
		RunCommand:   "echo Ran",
		RunOutput:    "output.txt",
	})
	assert.Nil(t, err)
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
	run.On("CreateStartEvent", "build").Return(10, nil)
	run.On("GetAppToDirectory", importPath).Return(nil)
	run.On("GetCacheToDirectory", cachePath).Return(nil)
	run.On("CreateLogEvent", "build", "stdout", "Build").Return(10, nil)
	run.On("CreateFinishEvent", "build", mock.Anything).Return(10, nil)
	run.On("PutCacheFromDirectory", cachePath).Return(nil)
	run.On("CreateStartEvent", "run").Return(10, nil)
	run.On("CreateLogEvent", "run", "stdout", "Run").Return(10, nil)
	run.On("CreateFinishEvent", "run", mock.Anything).Return(10, nil)
	run.On("CreateLastEvent").Return(10, nil)

	err := Run(run, &Options{
		ImportPath:   importPath,
		CachePath:    cachePath,
		AppPath:      appPath,
		EnvPath:      envPath,
		Environment:  map[string]string{"FOO": "bar"},
		BuildCommand: `echo Build`,
		RunCommand:   `echo Run`,
	})
	assert.Nil(t, err)
	// Check that environment files have been set up correctly
	b, _ := ioutil.ReadFile(filepath.Join(envPath, "FOO"))
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
	run.On("CreateStartEvent", "build").Return(10, nil)
	run.On("GetAppToDirectory", importPath).Return(nil).Run(func(args mock.Arguments) {
		copy.Copy("fixtures/scrapers/hello-world", importPath)
	})
	// Let the client know that there is no cache in this case
	run.On("GetCacheToDirectory", cachePath).Return(errors.New("404 Not Found"))
	run.On("CreateLogEvent", "build", "stderr", "bash: failing_command: command not found").Return(10, nil)
	run.On("CreateFinishEvent", "build", mock.MatchedBy(func(e protocol.ExitDataStage) bool {
		// Check that the exit codes are something sensible
		return e.ExitCode == 127 &&
			// The usage values are going to be a little different each time. So, the best we
			// can do for the moment is just check that they are not zero
			// Also not checking network usage because it will be non-zero when run under Linux and zero when run on OS X
			e.Usage.WallTime > 0 &&
			e.Usage.CPUTime > 0 &&
			e.Usage.MaxRSS > 0
	})).Return(10, nil)
	run.On("PutCacheFromDirectory", cachePath).Return(nil)
	run.On("CreateLastEvent").Return(10, nil)

	err := Run(run, &Options{
		ImportPath:   importPath,
		CachePath:    cachePath,
		AppPath:      appPath,
		EnvPath:      envPath,
		BuildCommand: `bash -c "failing_command"`,
		RunCommand:   "echo Ran",
		RunOutput:    "output.txt",
	})
	assert.Nil(t, err)
	run.AssertExpectations(t)
}

func TestFailingRun(t *testing.T) {
	appPath, importPath, cachePath, envPath := createTemporaryDirectories()
	defer os.RemoveAll(appPath)
	defer os.RemoveAll(importPath)
	defer os.RemoveAll(cachePath)
	defer os.RemoveAll(envPath)

	run := new(mocks.RunInterface)
	run.On("CreateStartEvent", "build").Return(10, nil)
	run.On("GetAppToDirectory", importPath).Return(nil).Run(func(args mock.Arguments) {
		copy.Copy("fixtures/scrapers/hello-world", importPath)
	})
	// Let the client know that there is no cache in this case
	run.On("GetCacheToDirectory", cachePath).Return(errors.New("404 Not Found"))
	run.On("CreateLogEvent", "build", "stdout", "build").Return(10, nil)
	run.On("CreateFinishEvent", "build", mock.MatchedBy(func(e protocol.ExitDataStage) bool {
		// Check that the exit codes are something sensible
		return e.ExitCode == 0 &&
			// The usage values are going to be a little different each time. So, the best we
			// can do for the moment is just check that they are not zero
			// Also not checking network usage because it will be non-zero when run under Linux and zero when run on OS X
			e.Usage.WallTime > 0 &&
			e.Usage.CPUTime > 0 &&
			e.Usage.MaxRSS > 0
	})).Return(10, nil)
	run.On("PutCacheFromDirectory", cachePath).Return(nil)
	run.On("CreateStartEvent", "run").Return(10, nil)
	run.On("CreateLogEvent", "run", "stderr", "bash: failing_command: command not found").Return(10, nil)
	run.On("PutOutputFromFile", filepath.Join(appPath, "output.txt")).Return(nil)
	run.On("CreateFinishEvent", "run", mock.MatchedBy(func(e protocol.ExitDataStage) bool {
		// Check that the exit codes are something sensible
		return e.ExitCode == 127 &&
			// The usage values are going to be a little different each time. So, the best we
			// can do for the moment is just check that they are not zero
			// Also not checking network usage because it will be non-zero when run under Linux and zero when run on OS X
			e.Usage.WallTime > 0 &&
			e.Usage.CPUTime > 0 &&
			e.Usage.MaxRSS > 0
	})).Return(10, nil)
	run.On("CreateLastEvent").Return(10, nil)

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
	assert.Nil(t, err)
	run.AssertExpectations(t)
}

func TestInternalError(t *testing.T) {
	appPath, importPath, cachePath, envPath := createTemporaryDirectories()
	defer os.RemoveAll(appPath)
	defer os.RemoveAll(importPath)
	defer os.RemoveAll(cachePath)
	defer os.RemoveAll(envPath)

	run := new(mocks.RunInterface)
	// Let's simulate an error with the blob storage. So, the wrapper is trying to
	// get the application and there's a problem.
	run.On("GetAppToDirectory", importPath).Return(errors.New("Something went wrong"))
	run.On("CreateLogEvent", "", "interr", "Internal error. The run will be automatically restarted.").Return(10, nil)

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
