package main

import (
	"io"
	"log"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type createResult struct {
	RunName  string `json:"run_name"`
	RunToken string `json:"run_token"`
}

type logMessage struct {
	// TODO: Make the stream an enum
	Log, Stream string
}

func getClientSet() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}

func commandCreate(scraperName string) (createResult, error) {
	clientset, err := getClientSet()
	if err != nil {
		return createResult{}, err
	}

	runName, runToken, err := createSecret(clientset, scraperName)

	createResult := createResult{
		RunName:  runName,
		RunToken: runToken,
	}
	return createResult, err
}

func commandGetApp(runName string, w io.Writer) error {
	return retrieveFromStore(storeAccess, runName, "app", "tgz", w)
}

func commandPutApp(reader io.Reader, objectSize int64, runName string) error {
	return saveToStore(storeAccess, reader, objectSize, runName, "app", "tgz")
}

func commandGetCache(runName string, w io.Writer) error {
	return retrieveFromStore(storeAccess, runName, "cache", "tgz", w)
}

func commandPutCache(reader io.Reader, objectSize int64, runName string) error {
	return saveToStore(storeAccess, reader, objectSize, runName, "cache", "tgz")
}

func commandGetOutput(runName string, w io.Writer) error {
	return retrieveFromStore(storeAccess, runName, "output", "", w)
}

func commandPutOutput(reader io.Reader, objectSize int64, runName string) error {
	return saveToStore(storeAccess, reader, objectSize, runName, "output", "")
}

func commandGetExitData(runName string, w io.Writer) error {
	return retrieveFromStore(storeAccess, runName, "exit-data", "json", w)
}

func commandPutExitData(reader io.Reader, objectSize int64, runName string) error {
	return saveToStore(storeAccess, reader, objectSize, runName, "exit-data", "json")
}

func commandStart(runName string, l startBody) error {
	clientset, err := getClientSet()
	if err != nil {
		return err
	}
	return createJob(clientset, runName, l)
}

func commandGetLogs(runName string) (io.ReadCloser, error) {
	clientset, err := getClientSet()
	if err != nil {
		return nil, err
	}
	return logStream(clientset, runName)
}

func commandCreateLog(runName string, l logMessage) error {
	// For the time being just show the results on stdout
	// TODO: Send them to the user with an http POST
	log.Printf("log %s %q", l.Stream, l.Log)
	return nil
}

func commandDelete(runName string) error {
	clientset, err := getClientSet()
	if err != nil {
		return err
	}

	err = deleteJob(clientset, runName)
	if err != nil {
		return err
	}
	err = deleteSecret(clientset, runName)
	if err != nil {
		return err
	}
	err = deleteFromStore(storeAccess, runName, "app", "tgz")
	if err != nil {
		return err
	}
	err = deleteFromStore(storeAccess, runName, "output", "")
	if err != nil {
		return err
	}
	err = deleteFromStore(storeAccess, runName, "exit-data", "json")
	if err != nil {
		return err
	}
	return deleteFromStore(storeAccess, runName, "cache", "tgz")
}

var storeAccess StoreAccess

func init() {
	var err error
	storeAccess, err = NewMinioAccess(
		// TODO: Get data store url for configmap
		"minio-service:9000",
		// TODO: Make bucket name configurable
		"clay",
		os.Getenv("STORE_ACCESS_KEY"),
		os.Getenv("STORE_SECRET_KEY"),
	)
	if err != nil {
		log.Fatal(err)
	}
}
