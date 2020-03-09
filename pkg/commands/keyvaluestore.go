package commands

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/openaustralia/yinyo/pkg/keyvaluestore"
)

func (app *AppImplementation) newCreatedKey(runID string) Key {
	return app.newKey(runID, "created")
}

func (app *AppImplementation) newCallbackKey(runID string) Key {
	return app.newKey(runID, "url")
}

func (app *AppImplementation) newExitDataKey(runID string, key string) Key {
	return app.newKey(runID, "exit_data/"+key)
}

func (app *AppImplementation) newExitDataFinishedKey(runID string) Key {
	return app.newExitDataKey(runID, "finished")
}

func (app *AppImplementation) newExitDataAPIKey(runID string, key string) Key {
	return app.newExitDataKey(runID, "api/"+key)
}

func (app *AppImplementation) newExitDataAPINetworkInKey(runID string) Key {
	return app.newExitDataAPIKey(runID, "network_in")
}

func (app *AppImplementation) newExitDataAPINetworkOutKey(runID string) Key {
	return app.newExitDataAPIKey(runID, "network_out")
}

func (app *AppImplementation) deleteAllKeys(runID string) error {
	// TODO: If one of these deletes fails just carry on
	err := app.newExitDataKey(runID, "build").delete()
	if err != nil {
		return err
	}
	err = app.newExitDataKey(runID, "run").delete()
	if err != nil {
		return err
	}
	err = app.newExitDataAPINetworkInKey(runID).delete()
	if err != nil {
		return err
	}
	err = app.newExitDataAPINetworkOutKey(runID).delete()
	if err != nil {
		return err
	}
	err = app.newExitDataFinishedKey(runID).delete()
	if err != nil {
		return err
	}
	err = app.newCallbackKey(runID).delete()
	if err != nil {
		return err
	}
	return app.newCreatedKey(runID).delete()
}

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
