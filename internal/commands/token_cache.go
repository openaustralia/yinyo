package commands

func tokenCacheKey(runName string) string {
	return "token:" + runName
}

func (app *AppImplementation) setTokenCache(runName string, runToken string) error {
	return app.KeyValueStore.Set(tokenCacheKey(runName), runToken)
}

// GetTokenCache gets the cached runToken
func (app *AppImplementation) GetTokenCache(runName string) (string, error) {
	return app.KeyValueStore.Get(tokenCacheKey(runName))
}

func (app *AppImplementation) deleteTokenCache(runName string) error {
	return app.KeyValueStore.Delete(tokenCacheKey(runName))
}
