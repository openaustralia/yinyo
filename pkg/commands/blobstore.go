package commands

import (
	"fmt"
	"io"
)

const filenameApp = "app.tgz"
const filenameCache = "cache.tgz"
const filenameOutput = "output"

func blobStoreStoragePath(runID string, fileName string) string {
	return runID + "/" + fileName
}

func (app *AppImplementation) getBlobStoreData(runID string, fileName string) (io.Reader, error) {
	p := blobStoreStoragePath(runID, fileName)
	r, err := app.BlobStore.Get(p)
	if err != nil && app.BlobStore.IsNotExist(err) {
		return r, fmt.Errorf("blobstore %v: %w", p, ErrNotFound)
	}
	return r, err
}

func (app *AppImplementation) putBlobStoreData(reader io.Reader, objectSize int64, runID string, fileName string) error {
	return app.BlobStore.Put(
		blobStoreStoragePath(runID, fileName),
		reader,
		objectSize,
	)
}

func (app *AppImplementation) deleteBlobStoreData(runID string, fileName string) error {
	return app.BlobStore.Delete(blobStoreStoragePath(runID, fileName))
}
