package blobstore

import (
	"io"
)

// Client defines the interface to access the storage layer
type Client interface {
	Put(path string, reader io.Reader, objectSize int64) error
	Get(path string) (io.Reader, error)
	Delete(path string) error
}
