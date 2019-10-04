package stream

// This is a distributed stream: something where we can add strings to a key
// and those strings can be streamed from one or more other places
// In our case we're implementing this using redis 5.0

type Stream interface {
	Add(key string, value string) error
	Get(key string, id string) (newId string, value string, finished bool, err error)
	Delete(key string) error
}
