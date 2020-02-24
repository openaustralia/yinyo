package commands

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/openaustralia/yinyo/pkg/keyvaluestore"
)

const tokenCacheKey = "token"
const callbackKey = "url"
const exitDataKeyBase = "exit_data/"
const exitDataFinishedKey = exitDataKeyBase + "finished"
const exitDataBuildKey = exitDataKeyBase + "build"
const exitDataRunKey = exitDataKeyBase + "run"
const exitDataAPIKeyBase = exitDataKeyBase + "api/"
const exitDataAPINetworkInKey = exitDataAPIKeyBase + "network_in"
const exitDataAPINetworkOutKey = exitDataAPIKeyBase + "network_out"

func keyValuePath(runName string, key string) string {
	return runName + "/" + key
}

func (app *AppImplementation) setKeyValueData(runName string, key string, value string) error {
	return app.KeyValueStore.Set(keyValuePath(runName, key), value)
}

func (app *AppImplementation) getKeyValueData(runName string, key string) (string, error) {
	value, err := app.KeyValueStore.Get(keyValuePath(runName, key))
	if errors.Is(err, keyvaluestore.ErrKeyNotExist) {
		return value, fmt.Errorf("%w", ErrNotFound)
	}
	return value, err
}

func (app *AppImplementation) getKeyValueDataAsInt(runName string, key string) (int64, error) {
	v, err := app.getKeyValueData(runName, key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(v, 10, 64)
}

func (app *AppImplementation) deleteKeyValueData(runName string, key string) error {
	return app.KeyValueStore.Delete(keyValuePath(runName, key))
}

func (app *AppImplementation) incrementKeyValueData(runName string, key string, value int64) (int64, error) {
	return app.KeyValueStore.Increment(keyValuePath(runName, key), value)
}
