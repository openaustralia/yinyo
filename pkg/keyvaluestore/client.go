package keyvaluestore

import "errors"

// KeyValueStore defines the interface to access the key value store
type KeyValueStore interface {
	Set(key string, value string) error
	Get(key string) (string, error)
	Delete(key string) error
	Increment(key string, value int64) (int64, error)
}

// ErrKeyNotExist is returned when a key doesn't exist
var ErrKeyNotExist = errors.New("key does not exist")
