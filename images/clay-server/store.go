package main

import (
	"io"

	"github.com/minio/minio-go/v6"
)

func saveToStore(reader io.Reader, objectSize int64, runName string, fileName string, fileExtension string) error {
	minioClient, err := minio.New(
		// TODO: Get access key and password from secret
		"minio-service:9000", "clay", "changeme123", false,
	)
	if err != nil {
		return err
	}

	path := fileName + "/" + runName
	if fileExtension != "" {
		path += "." + fileExtension
	}

	_, err = minioClient.PutObject(
		// TODO: Make bucket name configurable
		"clay",
		path,
		reader,
		objectSize,
		minio.PutObjectOptions{},
	)

	return err
}

func retrieveFromStore(runName string, fileName string, fileExtension string, writer io.Writer) error {
	minioClient, err := minio.New(
		// TODO: Get access key and password from secret
		"minio-service:9000", "clay", "changeme123", false,
	)
	if err != nil {
		return err
	}

	path := fileName + "/" + runName
	if fileExtension != "" {
		path += "." + fileExtension
	}

	// TODO: Make bucket name configurable
	object, err := minioClient.GetObject("clay", path, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, object)
	return err
}

func deleteFromStore(runName string, fileName string, fileExtension string) error {
	minioClient, err := minio.New(
		// TODO: Get access key and password from secret
		// TODO: This should only give it access to the one bucket
		"minio-service:9000", "clay", "changeme123", false,
	)
	if err != nil {
		return err
	}

	path := fileName + "/" + runName
	if fileExtension != "" {
		path += "." + fileExtension
	}

	err = minioClient.RemoveObject("clay", path)
	return err
}
