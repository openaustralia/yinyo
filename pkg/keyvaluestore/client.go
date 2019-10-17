package keyvaluestore

// Client defines the interface to access the key value store
type Client interface {
	Set(key string, value interface{}) error
	Get(key string) (interface{}, error)
	Delete(key string) error
}
