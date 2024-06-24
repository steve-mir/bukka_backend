package cache

import (
	"context"

	// "encoding/json"

	// "log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func NewCache(addr, username, password string, db int) *Cache {
	return &Cache{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
			Username: username,
		}),
	}
}

func (c *Cache) XAdd(ctx context.Context, stream, id string, values map[string]interface{}) (string, error) {
	args := &redis.XAddArgs{
		Stream:     stream,
		ID:         id, // You can use "*" to let Redis generate a unique ID
		Values:     values,
		MaxLen:     0,     // If you want to limit the length of the stream, set this to a positive number
		Approx:     true,  // Set to true to allow for approximate trimming of the stream to MaxLen
		NoMkStream: false, // Set to true if you don't want to create the stream if it doesn't exist
	}
	return c.client.XAdd(ctx, args).Result()
}

func (c *Cache) SetKey(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *Cache) XRead(ctx context.Context, streams []string, count int64, block time.Duration) ([]redis.XStream, error) {
	return c.client.XRead(ctx, &redis.XReadArgs{
		Streams: streams,
		Count:   count,
		Block:   block,
	}).Result()
}

func (c *Cache) GetKey(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *Cache) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

func (c *Cache) Decr(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, key).Result()
}

func (c *Cache) XGroupCreate(ctx context.Context, stream, group, start string) error {
	return c.client.XGroupCreateMkStream(ctx, stream, group, start).Err()
}

func (c *Cache) XAck(ctx context.Context, stream, group, id string) (int64, error) {
	return c.client.XAck(ctx, stream, group, id).Result()
}

func (c *Cache) DeleteKeyOld(ctx context.Context, key string) (int64, error) {
	return c.client.Del(ctx, key).Result()
}
func (c *Cache) DeleteKey(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// SIsMember checks if a member is in a Redis set.
func (c *Cache) SIsMember(ctx context.Context, setKey string, member interface{}) (bool, error) {
	cmd := c.client.SIsMember(ctx, setKey, member)
	return cmd.Result()
}

// SAdd adds a member to a Redis set.
func (c *Cache) SAdd(ctx context.Context, setKey string, member interface{}) error {
	cmd := c.client.SAdd(ctx, setKey, member)
	_, err := cmd.Result()
	return err
}

// SRem removes a member from a Redis set.
func (c *Cache) SRem(ctx context.Context, setKey string, member interface{}) error {
	cmd := c.client.SRem(ctx, setKey, member)
	_, err := cmd.Result()
	return err
}
