package main

import (
	"io"
)

func storagePath(runName string, fileName string) string {
	return runName + "/" + fileName
}

func saveToStore(m StoreAccess, reader io.Reader, objectSize int64, runName string, fileName string) error {
	return m.Put(
		storagePath(runName, fileName),
		reader,
		objectSize,
	)
}

func retrieveFromStore(m StoreAccess, runName string, fileName string, writer io.Writer) error {
	reader, err := m.Get(storagePath(runName, fileName))
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, reader)
	return err
}

func deleteFromStore(m StoreAccess, runName string, fileName string) error {
	return m.Delete(storagePath(runName, fileName))
}
