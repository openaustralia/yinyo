package stream

import (
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/openaustralia/yinyo/pkg/event"
)

type redisStream struct {
	client *redis.Client
}

// NewRedis returns the Redis implementation of Stream
func NewRedis(redisClient *redis.Client) Client {
	return &redisStream{client: redisClient}
}

func (stream *redisStream) add(key string, value string) (string, error) {
	return stream.client.XAdd(&redis.XAddArgs{
		Stream: key,
		Values: map[string]interface{}{"json": value},
	}).Result()
}

func (stream *redisStream) Add(key string, event event.Event) (addedEvent event.Event, err error) {
	b, err := json.Marshal(event)
	if err != nil {
		return
	}
	id, err := stream.add(key, string(b))
	if err != nil {
		return
	}
	addedEvent = event
	// Add the id to the returned event
	addedEvent.ID = id
	return
}

// Get the next string in the stream based on the id. It will wait until it's
// available
func (stream *redisStream) get(key string, id string) (newID string, value string, err error) {
	// For the moment get one event at a time
	// TODO: Grab more than one at a time for a little more efficiency
	result, err := stream.client.XRead(&redis.XReadArgs{
		Streams: []string{key, id},
		Count:   1,
		Block:   0,
	}).Result()
	if err != nil {
		return
	}
	newID = result[0].Messages[0].ID
	value = result[0].Messages[0].Values["json"].(string)
	return
}

func (stream *redisStream) Get(key string, id string) (event event.Event, err error) {
	newID, jsonString, err := stream.get(key, id)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(jsonString), &event)
	// Add the id to the event
	event.ID = newID
	return
}

func (stream *redisStream) Delete(key string) error {
	return stream.client.Del(key).Err()
}
