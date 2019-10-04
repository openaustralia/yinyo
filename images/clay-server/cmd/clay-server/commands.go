package main

import (
	"io"

	"github.com/dchest/uniuri"
	"github.com/go-redis/redis"

	"github.com/openaustralia/morph-ng/pkg/jobdispatcher"
	"github.com/openaustralia/morph-ng/pkg/store"
)

type createResult struct {
	RunName  string `json:"run_name"`
	RunToken string `json:"run_token"`
}

type logMessage struct {
	// TODO: Make the stream, stage and type an enum
	Log, Stream, Stage, Type string
}

func commandCreate(k jobdispatcher.Client, namePrefix string) (createResult, error) {
	if namePrefix == "" {
		namePrefix = "run"
	}
	// Generate random token
	runToken := uniuri.NewLen(32)
	runName, err := k.CreateJobAndToken(namePrefix, runToken)

	createResult := createResult{
		RunName:  runName,
		RunToken: runToken,
	}
	return createResult, err
}

func commandGetApp(storeAccess store.Client, runName string, w io.Writer) error {
	return retrieveFromStore(storeAccess, runName, "app.tgz", w)
}

func commandPutApp(storeAccess store.Client, reader io.Reader, objectSize int64, runName string) error {
	return saveToStore(storeAccess, reader, objectSize, runName, "app.tgz")
}

func commandGetCache(storeAccess store.Client, runName string, w io.Writer) error {
	return retrieveFromStore(storeAccess, runName, "cache.tgz", w)
}

func commandPutCache(storeAccess store.Client, reader io.Reader, objectSize int64, runName string) error {
	return saveToStore(storeAccess, reader, objectSize, runName, "cache.tgz")
}

func commandGetOutput(storeAccess store.Client, runName string, w io.Writer) error {
	return retrieveFromStore(storeAccess, runName, "output", w)
}

func commandPutOutput(storeAccess store.Client, reader io.Reader, objectSize int64, runName string) error {
	return saveToStore(storeAccess, reader, objectSize, runName, "output")
}

func commandGetExitData(storeAccess store.Client, runName string, w io.Writer) error {
	return retrieveFromStore(storeAccess, runName, "exit-data.json", w)
}

func commandPutExitData(storeAccess store.Client, reader io.Reader, objectSize int64, runName string) error {
	return saveToStore(storeAccess, reader, objectSize, runName, "exit-data.json")
}

func commandStart(k jobdispatcher.Client, runName string, output string, env map[string]string) error {
	return k.StartJob(runName, "openaustralia/clay-scraper:v1", []string{"/bin/run.sh", runName, output}, env)
}

func commandGetEvent(redisClient *redis.Client, runName string, id string) (newId string, jsonString string, finished bool, err error) {
	// For the moment get one event at a time
	// TODO: Grab more than one at a time for a little more efficiency
	result, err := redisClient.XRead(&redis.XReadArgs{
		Streams: []string{runName, id},
		Count:   1,
		Block:   0,
	}).Result()
	if err != nil {
		return
	}
	newId = result[0].Messages[0].ID
	jsonString = result[0].Messages[0].Values["json"].(string)

	if jsonString == "EOF" {
		finished = true
	}
	return
}

func commandCreateEvent(redisClient *redis.Client, runName string, eventJson string) error {
	// TODO: Send the event to the user with an http POST

	// Send the json to a redis stream
	return redisClient.XAdd(&redis.XAddArgs{
		// TODO: Use something like runName-events instead for the stream name
		Stream: runName,
		Values: map[string]interface{}{"json": eventJson},
	}).Err()
}

func commandDelete(k jobdispatcher.Client, storeAccess store.Client, redisClient *redis.Client, runName string) error {
	err := k.DeleteJobAndToken(runName)
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
