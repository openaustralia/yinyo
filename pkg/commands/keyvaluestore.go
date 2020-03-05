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

type Key struct {
	key    string
	client keyvaluestore.KeyValueStore
}

func (app *AppImplementation) newKey(runID string, key string) Key {
	return Key{key: runID + "/" + key, client: app.KeyValueStore}
}

func (key Key) set(value string) error {
	return key.client.Set(key.key, value)
}

func (key Key) get() (string, error) {
	value, err := key.client.Get(key.key)
	if errors.Is(err, keyvaluestore.ErrKeyNotExist) {
		return value, fmt.Errorf("%w", ErrNotFound)
	}
	return value, err
}

func (key Key) getAsInt() (int64, error) {
	v, err := key.get()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(v, 10, 64)
}

func (key Key) delete() error {
	return key.client.Delete(key.key)
}

func (key Key) increment(value int64) (int64, error) {
	return key.client.Increment(key.key, value)
}
