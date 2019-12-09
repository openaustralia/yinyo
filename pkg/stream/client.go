package stream

import "github.com/openaustralia/yinyo/pkg/event"

// This is a distributed stream: something where we can add events to a key
// and those events can be streamed from one or more other places
// In our case we're implementing this using redis 5.0

// Client is the interface for accessing the distributed stream
type Client interface {
	Add(key string, event event.Event) (addedEvent event.Event, err error)
	Get(key string, id string) (event event.Event, err error)
	Delete(key string) error
}
