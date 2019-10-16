package keyvaluestore

import (
	"github.com/go-redis/redis"
)

type client struct {
	client *redis.Client
}

// NewRedis returns the Redis implementation of Client
func NewRedis(redisClient *redis.Client) Client {
	return &client{client: redisClient}
}

func namespaced(key string) string {
	return "kv:" + key
}

func (client *client) Set(key string, value interface{}) error {
	// TODO: Do we want to set an expiration here? If so what and how does it know
	// the correct value?
	return client.client.Set(namespaced(key), value, 0).Err()
}

func (client *client) Get(key string) (interface{}, error) {
	return client.client.Get(namespaced(key)).Result()
}

func (client *client) Delete(key string) error {
	return client.client.Del(namespaced(key)).Err()
}
