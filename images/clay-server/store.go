package main

import (
	"io"
	"os"

	"github.com/minio/minio-go/v6"
)

func minioClient() (*minio.Client, error) {
	return minio.New(
		// TODO: Get data store url for configmap
		"minio-service:9000",
		os.Getenv("STORE_ACCESS_KEY"),
		os.Getenv("STORE_SECRET_KEY"),
		false,
	)
}

func storagePath(runName string, fileName string, fileExtension string) string {
	path := fileName + "/" + runName
	if fileExtension != "" {
		path += "." + fileExtension
	}
	return path
}

func saveToStore(reader io.Reader, objectSize int64, runName string, fileName string, fileExtension string) error {
	minioClient, err := minioClient()
	if err != nil {
		return err
	}

	_, err = minioClient.PutObject(
		// TODO: Make bucket name configurable
		"clay",
		storagePath(runName, fileName, fileExtension),
		reader,
		objectSize,
		minio.PutObjectOptions{},
	)

	return err
}

func retrieveFromStore(runName string, fileName string, fileExtension string, writer io.Writer) error {
	minioClient, err := minioClient()
	if err != nil {
		return err
	}

	// TODO: Make bucket name configurable
	object, err := minioClient.GetObject(
		"clay",
		storagePath(runName, fileName, fileExtension),
		minio.GetObjectOptions{},
	)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, object)
	return err
}

func deleteFromStore(runName string, fileName string, fileExtension string) error {
	minioClient, err := minioClient()
	if err != nil {
		return err
	}

	err = minioClient.RemoveObject(
		"clay",
		storagePath(runName, fileName, fileExtension),
	)
	return err
}
