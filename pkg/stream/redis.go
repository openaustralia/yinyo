package stream

import (
	"github.com/go-redis/redis"
)

type redisStream struct {
	client *redis.Client
}

// NewRedis returns the Redis implementation of Stream
func NewRedis(redisClient *redis.Client) Client {
	return &redisStream{client: redisClient}
}

func (stream *redisStream) Add(key string, value string) error {
	return stream.client.XAdd(&redis.XAddArgs{
		Stream: key,
		Values: map[string]interface{}{"json": value},
	}).Err()
}

// Get the next string in the stream based on the id. It will wait until it's
// available
func (stream *redisStream) Get(key string, id string) (newID string, value string, err error) {
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

func (stream *redisStream) Delete(key string) error {
	return stream.client.Del(key).Err()
}
