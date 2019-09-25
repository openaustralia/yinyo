package main

import (
	"io"
	"log"

	"github.com/go-redis/redis"
	"k8s.io/client-go/kubernetes"
)

type createResult struct {
	RunName  string `json:"run_name"`
	RunToken string `json:"run_token"`
}

type logMessage struct {
	// TODO: Make the stream and stage an enum
	Log, Stream, Stage string
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

func commandCreateLog(redisClient *redis.Client, runName string, l string) error {
	// For the time being just show the results on stdout
	// TODO: Send them to the user with an http POST
	log.Println(l)

	// Send the json to a redis stream
	return redisClient.XAdd(&redis.XAddArgs{
		// TODO: Use something like runName-events instead for the stream name
		Stream: runName,
		Values: map[string]interface{}{"json": l},
	}).Err()
}

func commandDelete(clientset *kubernetes.Clientset, storeAccess StoreAccess, redisClient *redis.Client, runName string) error {
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
	err = deleteFromStore(storeAccess, runName, "cache.tgz")
	if err != nil {
		return err
	}
	return redisClient.Del(runName).Err()
}
