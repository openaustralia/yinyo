package commands

import (
	"encoding/json"
	"errors"
	"fmt"

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

func (app *AppImplementation) newExitDataExitCodeKey(runID string, stage string) Key {
	return app.newExitDataKey(runID, stage+"/exit_code")
}

func (app *AppImplementation) newExitDataMaxRSSKey(runID string, stage string) Key {
	return app.newExitDataKey(runID, stage+"/max_rss")
}

func (app *AppImplementation) newExitDataNetworkInKey(runID string, stage string) Key {
	return app.newExitDataKey(runID, stage+"/network_in")
}

func (app *AppImplementation) newExitDataNetworkOutKey(runID string, stage string) Key {
	return app.newExitDataKey(runID, stage+"/network_out")
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

func (app *AppImplementation) deleteExitDataKeys(runID string, stage string) error {
	err := app.newExitDataExitCodeKey(runID, stage).delete()
	if err != nil {
		return err
	}

	err = app.newExitDataMaxRSSKey(runID, stage).delete()
	if err != nil {
		return err
	}

	err = app.newExitDataNetworkInKey(runID, stage).delete()
	if err != nil {
		return err
	}

	err = app.newExitDataNetworkOutKey(runID, stage).delete()
	if err != nil {
		return err
	}

	return nil
}

func (app *AppImplementation) deleteAllKeys(runID string) error {
	// TODO: If one of these deletes fails just carry on
	err := app.deleteExitDataKeys(runID, "build")
	if err != nil {
		return err
	}
	err = app.deleteExitDataKeys(runID, "run")
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

func (key Key) setAsInt(value int) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return key.set(string(b))
}

func (key Key) setAsUint64(value uint64) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return key.set(string(b))
}

func (key Key) get() (string, error) {
	value, err := key.client.Get(key.key)
	if errors.Is(err, keyvaluestore.ErrKeyNotExist) {
		return value, fmt.Errorf("%w", ErrNotFound)
	}
	return value, err
}

func (key Key) getAsInt() (int, error) {
	var value int
	string, err := key.get()
	if err != nil {
		return value, err
	}
	err = json.Unmarshal([]byte(string), &value)
	return value, err
}

func (key Key) getAsInt64() (int64, error) {
	var value int64
	string, err := key.get()
	if err != nil {
		return value, err
	}
	err = json.Unmarshal([]byte(string), &value)
	return value, err
}

func (key Key) getAsUint64() (uint64, error) {
	var value uint64
	string, err := key.get()
	if err != nil {
		return value, err
	}
	err = json.Unmarshal([]byte(string), &value)
	return value, err
}

func (key Key) delete() error {
	return key.client.Delete(key.key)
}

func (key Key) increment(value int64) (int64, error) {
	return key.client.Increment(key.key, value)
}
