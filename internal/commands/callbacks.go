package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/openaustralia/yinyo/pkg/event"
)

func callbackURLKey(runName string) string {
	return "url:" + runName
}

func (app *App) setCallbackURL(runName string, callbackURL string) error {
	return app.KeyValueStore.Set(callbackURLKey(runName), callbackURL)
}

func (app *App) getCallbackURL(runName string) (string, error) {
	r, err := app.KeyValueStore.Get(callbackURLKey(runName))
	if err != nil {
		return "", err
	}
	callbackURL, ok := r.(string)
	if !ok {
		return "", errors.New("Unexpected type")
	}
	return callbackURL, nil
}

func (app *App) deleteCallbackURL(runName string) error {
	return app.KeyValueStore.Delete(callbackURLKey(runName))
}

func (app *App) postCallbackEvent(runName string, event event.Event) error {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.Encode(event)

	callbackURL, err := app.getCallbackURL(runName)
	if err != nil {
		return err
	}

	// Only do the callback if there's a sensible URL
	if callbackURL != "" {
		resp, err := app.HTTP.Post(callbackURL, "application/json", &b)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return errors.New("callback: " + resp.Status)
		}
	}
	return nil
}
