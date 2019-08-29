package main

import (
	"io"
)

func storagePath(runName string, fileName string, fileExtension string) string {
	path := runName + "/" + fileName
	if fileExtension != "" {
		path += "." + fileExtension
	}
	return path
}

func saveToStore(m StoreAccess, reader io.Reader, objectSize int64, runName string, fileName string, fileExtension string) error {
	return m.Put(
		storagePath(runName, fileName, fileExtension),
		reader,
		objectSize,
	)
}

func retrieveFromStore(m StoreAccess, runName string, fileName string, fileExtension string, writer io.Writer) error {
	reader, err := m.Get(storagePath(runName, fileName, fileExtension))
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, reader)
	return err
}

func deleteFromStore(m StoreAccess, runName string, fileName string, fileExtension string) error {
	return m.Delete(storagePath(runName, fileName, fileExtension))
}
