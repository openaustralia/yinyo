package commands

import (
	"fmt"
	"io"
)

const filenameApp = "app.tgz"
const filenameCache = "cache.tgz"
const filenameOutput = "output"

func blobStoreStoragePath(runName string, fileName string) string {
	return runName + "/" + fileName
}

func (app *AppImplementation) getBlobStoreData(runName string, fileName string) (io.Reader, error) {
	p := blobStoreStoragePath(runName, fileName)
	r, err := app.BlobStore.Get(p)
	if err != nil && app.BlobStore.IsNotExist(err) {
		return r, fmt.Errorf("blobstore %v: %w", p, ErrNotFound)
	}
	return r, err
}

func (app *AppImplementation) putBlobStoreData(reader io.Reader, objectSize int64, runName string, fileName string) error {
	return app.BlobStore.Put(
		blobStoreStoragePath(runName, fileName),
		reader,
		objectSize,
	)
}

func (app *AppImplementation) deleteBlobStoreData(runName string, fileName string) error {
	return app.BlobStore.Delete(blobStoreStoragePath(runName, fileName))
}
