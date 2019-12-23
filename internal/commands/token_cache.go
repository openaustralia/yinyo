package commands

import (
	"errors"
	"fmt"

	"github.com/openaustralia/yinyo/pkg/keyvaluestore"
)

func tokenCacheKey(runName string) string {
	return "token:" + runName
}

func (app *AppImplementation) setTokenCache(runName string, runToken string) error {
	return app.KeyValueStore.Set(tokenCacheKey(runName), runToken)
}

// GetTokenCache gets the cached runToken. Returns ErrNotFound if run name doesn't exist
func (app *AppImplementation) GetTokenCache(runName string) (string, error) {
	value, err := app.KeyValueStore.Get(tokenCacheKey(runName))
	if err != nil {
		if errors.Is(err, keyvaluestore.ErrKeyNotExist) {
			return value, fmt.Errorf("%w", ErrNotFound)
		}
		return value, err
	}
	return value, nil
}

func (app *AppImplementation) deleteTokenCache(runName string) error {
	return app.KeyValueStore.Delete(tokenCacheKey(runName))
}
