package commands

import (
	"errors"
	"fmt"

	"github.com/openaustralia/yinyo/pkg/keyvaluestore"
)

const tokenCacheKey = "token"
const callbackKey = "url"

func (app *AppImplementation) setKeyValueData(runName string, key string, value string) error {
	// TODO: Reverse order of key and runName
	return app.KeyValueStore.Set(key+":"+runName, value)
}

func (app *AppImplementation) getKeyValueData(runName string, key string) (string, error) {
	value, err := app.KeyValueStore.Get(key + ":" + runName)
	if errors.Is(err, keyvaluestore.ErrKeyNotExist) {
		return value, fmt.Errorf("%w", ErrNotFound)
	}
	return value, err
}

func (app *AppImplementation) deleteKeyValueData(runName string, key string) error {
	return app.KeyValueStore.Delete(key + ":" + runName)
}
