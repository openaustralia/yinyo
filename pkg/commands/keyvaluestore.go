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
