package commands

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/openaustralia/yinyo/pkg/keyvaluestore"
)

const createdKey = "created"
const callbackKey = "url"
const exitDataKeyBase = "exit_data/"
const exitDataFinishedKey = exitDataKeyBase + "finished"
const exitDataBuildKey = exitDataKeyBase + "build"
const exitDataRunKey = exitDataKeyBase + "run"
const exitDataAPIKeyBase = exitDataKeyBase + "api/"
const exitDataAPINetworkInKey = exitDataAPIKeyBase + "network_in"
const exitDataAPINetworkOutKey = exitDataAPIKeyBase + "network_out"

func keyValuePath(runID string, key string) string {
	return runID + "/" + key
}

func (app *AppImplementation) setKeyValueData(runID string, key string, value string) error {
	return app.KeyValueStore.Set(keyValuePath(runID, key), value)
}

func (app *AppImplementation) getKeyValueData(runID string, key string) (string, error) {
	value, err := app.KeyValueStore.Get(keyValuePath(runID, key))
	if errors.Is(err, keyvaluestore.ErrKeyNotExist) {
		return value, fmt.Errorf("%w", ErrNotFound)
	}
	return value, err
}

func (app *AppImplementation) getKeyValueDataAsInt(runID string, key string) (int64, error) {
	v, err := app.getKeyValueData(runID, key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(v, 10, 64)
}

func (app *AppImplementation) deleteKeyValueData(runID string, key string) error {
	return app.KeyValueStore.Delete(keyValuePath(runID, key))
}

func (app *AppImplementation) incrementKeyValueData(runID string, key string, value int64) (int64, error) {
	return app.KeyValueStore.Increment(keyValuePath(runID, key), value)
}
