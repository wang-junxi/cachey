// Package cachey provides a simple and flexible way to cache the results of functions
// using either redis or memory as the backend.
package cachey

import (
	"github.com/go-redis/redis/v8"
	"github.com/patrickmn/go-cache"
)

// Client represents a cachey client that can use either redis or memory as the cache backend.
type Client struct {
	redisClient  *redis.Client
	memoryClient *cache.Cache
}

// New creates a new cachey client with the given redis and memory clients.
func New(rc *redis.Client, cc *cache.Cache) *Client {
	return &Client{
		redisClient:  rc,
		memoryClient: cc,
	}
}

// M returns a new request that uses memory as the cache backend.
func (c *Client) M() *Request { return &Request{client: c, use: MemoryClient} }

// R returns a new request that uses redis as the cache backend.
func (c *Client) R() *Request { return &Request{client: c, use: RedisClient} }
