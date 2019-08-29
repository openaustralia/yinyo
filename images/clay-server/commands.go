package main

import (
	"io"
	"log"

	"k8s.io/client-go/kubernetes"
)

type createResult struct {
	RunName  string `json:"run_name"`
	RunToken string `json:"run_token"`
}

type logMessage struct {
	// TODO: Make the stream an enum
	Log, Stream string
}

func commandCreate(clientset *kubernetes.Clientset, scraperName string) (createResult, error) {
	runName, runToken, err := createSecret(clientset, scraperName)

	createResult := createResult{
		RunName:  runName,
		RunToken: runToken,
	}
	return createResult, err
}

func commandGetApp(storeAccess StoreAccess, runName string, w io.Writer) error {
	return retrieveFromStore(storeAccess, runName, "app.tgz", w)
}

func commandPutApp(storeAccess StoreAccess, reader io.Reader, objectSize int64, runName string) error {
	return saveToStore(storeAccess, reader, objectSize, runName, "app.tgz")
}

func commandGetCache(storeAccess StoreAccess, runName string, w io.Writer) error {
	return retrieveFromStore(storeAccess, runName, "cache.tgz", w)
}

func commandPutCache(storeAccess StoreAccess, reader io.Reader, objectSize int64, runName string) error {
	return saveToStore(storeAccess, reader, objectSize, runName, "cache.tgz")
}

func commandGetOutput(storeAccess StoreAccess, runName string, w io.Writer) error {
	return retrieveFromStore(storeAccess, runName, "output", w)
}

func commandPutOutput(storeAccess StoreAccess, reader io.Reader, objectSize int64, runName string) error {
	return saveToStore(storeAccess, reader, objectSize, runName, "output")
}

func commandGetExitData(storeAccess StoreAccess, runName string, w io.Writer) error {
	return retrieveFromStore(storeAccess, runName, "exit-data.json", w)
}

func commandPutExitData(storeAccess StoreAccess, reader io.Reader, objectSize int64, runName string) error {
	return saveToStore(storeAccess, reader, objectSize, runName, "exit-data.json")
}

func commandStart(clientset *kubernetes.Clientset, runName string, l startBody) error {
	return createJob(clientset, runName, l)
}

func commandGetLogs(clientset *kubernetes.Clientset, runName string) (io.ReadCloser, error) {
	return logStream(clientset, runName)
}

func commandCreateLog(runName string, l logMessage) error {
	// For the time being just show the results on stdout
	// TODO: Send them to the user with an http POST
	log.Printf("log %s %q", l.Stream, l.Log)
	return nil
}

func commandDelete(clientset *kubernetes.Clientset, storeAccess StoreAccess, runName string) error {
	err := deleteJob(clientset, runName)
	if err != nil {
		return err
	}
	err = deleteSecret(clientset, runName)
	if err != nil {
		return err
	}
	err = deleteFromStore(storeAccess, runName, "app.tgz")
	if err != nil {
		return err
	}
	err = deleteFromStore(storeAccess, runName, "output")
	if err != nil {
		return err
	}
	err = deleteFromStore(storeAccess, runName, "exit-data.json")
	if err != nil {
		return err
	}
	return deleteFromStore(storeAccess, runName, "cache.tgz")
}
