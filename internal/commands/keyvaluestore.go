package commands

import (
	"errors"
	"fmt"

	"github.com/openaustralia/yinyo/pkg/keyvaluestore"
)

const tokenCacheKey = "token"
const callbackKey = "url"
const exitDataKey = "exit_data"

func keyValuePath(runName string, key string) string {
	// TODO: Reverse order of key and runName
	return key + ":" + runName
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

func (app *AppImplementation) deleteKeyValueData(runName string, key string) error {
	return app.KeyValueStore.Delete(keyValuePath(runName, key))
}
