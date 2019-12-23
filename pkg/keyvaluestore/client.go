package keyvaluestore

import "errors"

// Client defines the interface to access the key value store
type Client interface {
	Set(key string, value string) error
	Get(key string) (string, error)
	Delete(key string) error
}

// ErrKeyNotExist is returned when a key doesn't exist
var ErrKeyNotExist = errors.New("key does not exist")
