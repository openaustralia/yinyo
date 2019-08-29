package main

import (
	"io"
	"os"

	"github.com/minio/minio-go/v6"
)

// MinioAccess implements StoreAccess
type MinioAccess struct {
	Client     *minio.Client
	BucketName string
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

func access() (MinioAccess, error) {
	minioClient, err := minio.New(
		// TODO: Get data store url for configmap
		"minio-service:9000",
		os.Getenv("STORE_ACCESS_KEY"),
		os.Getenv("STORE_SECRET_KEY"),
		false,
	)
	m := MinioAccess{
		Client: minioClient,
		// TODO: Make bucket name configurable
		BucketName: "clay",
	}
	return m, err
}

func storagePath(runName string, fileName string, fileExtension string) string {
	path := fileName + "/" + runName
	if fileExtension != "" {
		path += "." + fileExtension
	}
	return path
}

func saveToStore(reader io.Reader, objectSize int64, runName string, fileName string, fileExtension string) error {
	m, err := access()
	if err != nil {
		return err
	}
	return m.Put(
		storagePath(runName, fileName, fileExtension),
		reader,
		objectSize,
	)
}

func retrieveFromStore(runName string, fileName string, fileExtension string, writer io.Writer) error {
	m, err := access()
	if err != nil {
		return err
	}
	reader, err := m.Get(storagePath(runName, fileName, fileExtension))
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, reader)
	return err
}

func deleteFromStore(runName string, fileName string, fileExtension string) error {
	m, err := access()
	if err != nil {
		return err
	}
	return m.Delete(storagePath(runName, fileName, fileExtension))
}
