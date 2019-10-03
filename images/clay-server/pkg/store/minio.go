package store

import (
	"io"

	"github.com/minio/minio-go/v6"
)

type minioClient struct {
	Client     *minio.Client
	BucketName string
}

// NewMinioClient creates a MinioClient
func NewMinioClient(url string, bucketName string, accessKey string, secretKey string) (Client, error) {
	client, err := minio.New(url, accessKey, secretKey, false)
	m := &minioClient{
		Client:     client,
		BucketName: bucketName,
	}
	return m, err
}

// Put saves a file to the store with the given path
func (m *minioClient) Put(path string, reader io.Reader, objectSize int64) error {
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
func (m *minioClient) Get(path string) (io.Reader, error) {
	return m.Client.GetObject(
		m.BucketName,
		path,
		minio.GetObjectOptions{},
	)
}

// Delete removes a file in the store at the given path
func (m *minioClient) Delete(path string) error {
	return m.Client.RemoveObject(
		m.BucketName,
		path,
	)
}
