package keyvaluestore

// Client defines the interface to access the key value store
type Client interface {
	Set(key string, value string) error
	Get(key string) (string, error)
	Delete(key string) error
}
