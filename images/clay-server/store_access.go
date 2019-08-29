package main

import (
	"io"

	"github.com/minio/minio-go/v6"
)

// StoreAccess defines the interface to access the storage layer
type StoreAccess interface {
	Put(path string, reader io.Reader, objectSize int64) error
	Get(path string) (io.Reader, error)
	Delete(path string) error
}

// MinioAccess implements StoreAccess
type MinioAccess struct {
	Client     *minio.Client
	BucketName string
}

// NewMinioAccess creates a MinioAccess
func NewMinioAccess(url string, bucketName string, accessKey string, secretKey string) (StoreAccess, error) {
	client, err := minio.New(url, accessKey, secretKey, false)
	m := &MinioAccess{
		Client:     client,
		BucketName: bucketName,
	}
	return m, err
}

// Put saves a file to the store with the given path
func (m *MinioAccess) Put(path string, reader io.Reader, objectSize int64) error {
	_, err := m.Client.PutObject(
		m.BucketName,
		path,
		reader,
		objectSize,
		minio.PutObjectOptions{},
	)
	return err
}

// Get retrieves a file at the given path from the store
func (m *MinioAccess) Get(path string) (io.Reader, error) {
	return m.Client.GetObject(
		m.BucketName,
		path,
		minio.GetObjectOptions{},
	)
}

// Delete removes a file in the store at the given path
func (m *MinioAccess) Delete(path string) error {
	return m.Client.RemoveObject(
		m.BucketName,
		path,
	)
}
