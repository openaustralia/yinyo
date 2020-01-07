package blobstore

import (
	"io"
)

// BlobStore defines the interface to access the storage layer
type BlobStore interface {
	Put(path string, reader io.Reader, objectSize int64) error
	Get(path string) (io.Reader, error)
	Delete(path string) error
	IsNotExist(error) bool
}
