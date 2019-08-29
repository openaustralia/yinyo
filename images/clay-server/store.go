package main

import (
	"io"
	"os"
)

func access() (StoreAccess, error) {
	return NewMinioAccess(
		// TODO: Get data store url for configmap
		"minio-service:9000",
		// TODO: Make bucket name configurable
		"clay",
		os.Getenv("STORE_ACCESS_KEY"),
		os.Getenv("STORE_SECRET_KEY"),
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
