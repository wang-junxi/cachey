// Package cachey provides a simple and flexible way to cache the results of functions
// using either redis or memory as the backend.
package cachey

import (
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog"
)

// Client represents a cachey client that can use either redis or memory as the cache backend.
type Client struct {
	redisClient  *redis.Client
	memoryClient *cache.Cache
	logger       *zerolog.Logger
}

// New creates a new cachey client with the given redis and memory clients.
func New(rc *redis.Client, cc *cache.Cache) *Client {
	warnLogger := zerolog.New(os.Stdout).Level(zerolog.WarnLevel).With().Timestamp().Logger()
	return &Client{
		redisClient:  rc,
		memoryClient: cc,
		logger:       &warnLogger,
	}
}

// EnableDebug enable debugging more detailed message with zerolog.
func (r *Client) EnableDebug() *Client {
	debugLogger := r.logger.Level(zerolog.DebugLevel)
	r.logger = &debugLogger
	return r
}

// M returns a new request that uses memory as the cache backend.
func (c *Client) M() *Request { return &Request{client: c, use: MemoryClient} }

// R returns a new request that uses redis as the cache backend.
func (c *Client) R() *Request { return &Request{client: c, use: RedisClient} }
