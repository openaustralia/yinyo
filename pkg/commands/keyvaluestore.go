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

func (app *AppImplementation) newExitDataFinishedKey(runID string) Key {
	return app.newExitDataKey(runID, "finished")
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

func (key Key) set(value interface{}) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return key.client.Set(key.key, string(b))
}

func (key Key) get(value interface{}) error {
	string, err := key.client.Get(key.key)
	if err != nil {
		if errors.Is(err, keyvaluestore.ErrKeyNotExist) {
			return fmt.Errorf("%w", ErrNotFound)
		}
		return err
	}
	return json.Unmarshal([]byte(string), value)
}

func (key Key) delete() error {
	return key.client.Delete(key.key)
}
