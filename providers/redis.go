package providers

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/knadh/koanf/v2"
	"github.com/redis/go-redis/v9"
)

var _ koanf.Provider = (*Redis)(nil)

// Redis is an implementation of koanf.Provider that reads/loads configuration
// stored in Redis as a STRING type. Redis is capable of watching a key in Redis
// and notifying changes via a callback.
type Redis struct {
	client     *redis.Client
	key        string
	watched    atomic.Uint32
	pubsub     *redis.PubSub
	changeChan <-chan *redis.Message
}

// RedisProvider initializes and returns a new instance if Redis.
func RedisProvider(client *redis.Client, key string) *Redis {
	return &Redis{
		client:     client,
		key:        key,
		watched:    atomic.Uint32{},
		pubsub:     nil,
		changeChan: nil,
	}
}

// ReadBytes retrieves the configuration from Redis for the configured key.
func (r *Redis) ReadBytes() ([]byte, error) {
	data, err := r.client.Get(context.Background(), r.key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("key %s does not exist", r.key)
		}
		return nil, err
	}
	return data, nil
}

// Read is not supported by Redis and will always return an error.
func (r *Redis) Read() (map[string]interface{}, error) {
	return nil, fmt.Errorf("%T does not support Read()", r)
}

// Watch utilizes Redis keyspace events to detect when a key has been modified
// and invokes the callback.
//
// Since Watch relies on Redis keyspace events ensure it is enabled in the Redis
// or Watch will not behave as expected.
func (r *Redis) Watch(cb func(event interface{}, err error)) error {
	activated := r.watched.CompareAndSwap(0, 1)
	if !activated {
		return fmt.Errorf("%T.Watch may only be invoked once", r)
	}

	// Subscribe to keyspace notifications for changes to the specified key
	r.pubsub = r.client.PSubscribe(context.Background(), "__keyspace@0__:"+r.key)
	r.changeChan = r.pubsub.Channel()

	go func() {
		for msg := range r.changeChan {
			if msg.Payload == "del" {
				cb(nil, fmt.Errorf("configuration key deleted"))
				continue
			}

			cb(msg.Payload, nil)
		}
	}()

	return nil
}

// Close cleans up any resources and stops the watch if one was active.
func (r *Redis) Close() error {
	if r.watched.Load() == 1 && r.pubsub != nil {
		return r.pubsub.Close()
	}
	return nil
}
