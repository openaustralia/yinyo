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

func saveToStore(reader io.Reader, objectSize int64, runName string, fileName string, fileExtension string) error {
	minioClient, err := minioClient()
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
	minioClient, err := minioClient()
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
	minioClient, err := minioClient()
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
