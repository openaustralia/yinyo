package commands

import (
	"errors"
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

