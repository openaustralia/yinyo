package apiclient

// Utilities for easier handling of archives

import (
	"io"
	"os"

	"github.com/openaustralia/yinyo/pkg/archive"
)

// GetAppToDirectory downloads the scraper code into a pre-existing directory on the filesystem
func (run *Run) GetAppToDirectory(dir string) error {
	app, err := run.GetApp()
	if err != nil {
		return err
	}
	defer app.Close()
	return archive.ExtractToDirectory(app, dir)
}

// GetCacheToDirectory downloads the cache into a pre-existing directory on the filesystem
func (run *Run) GetCacheToDirectory(dir string) error {
	app, err := run.GetCache()
	if err != nil {
		return err
	}
	defer app.Close()
	return archive.ExtractToDirectory(app, dir)
}

// PutAppFromDirectory uploads the scraper code from a directory on the filesystem
// ignorePaths is a list of paths (relative to dir) that should be ignored and not uploaded
func (run *Run) PutAppFromDirectory(dir string, ignorePaths []string) error {
	r, err := archive.CreateFromDirectory(dir, ignorePaths)
	if err != nil {
		return err
	}
	return run.PutApp(r)
}

// PutCacheFromDirectory uploads the cache from a directory on the filesystem
func (run *Run) PutCacheFromDirectory(dir string) error {
	r, err := archive.CreateFromDirectory(dir, []string{})
	if err != nil {
		return err
	}
	return run.PutCache(r)
}

// GetOutputToFile downloads the output of the run and saves it in a file which it
// will create or overwrite.
func (run *Run) GetOutputToFile(path string) error {
	output, err := run.GetOutput()
	if err != nil {
		return err
	}
	defer output.Close()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, output)
	return err
}

// PutOutputFromFile uploads the contents of a file as the output of the scraper
func (run *Run) PutOutputFromFile(path string) error {
	// TODO: Don't do a separate Stat and Open
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		return run.PutOutput(f)
	}
	// We get here if output file doesn't exist. In that case we just want
	// to happily carry on like nothing weird has happened
	return nil
}

// GetCacheToFile downloads the cache (as a tar & gzipped file) and saves it (without uncompressing it)
func (run *Run) GetCacheToFile(path string) error {
	cache, err := run.GetCache()
	if err != nil {
		return err
	}
	defer cache.Close()

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, cache)
	return err
}
