package cachey

import (
	"github.com/go-redis/redis/v8"
	"github.com/patrickmn/go-cache"
)

type Client struct {
	redisClient  *redis.Client
	memoryClient *cache.Cache
}

func New(rc *redis.Client, cc *cache.Cache) *Client {
	return &Client{
		redisClient:  rc,
		memoryClient: cc,
	}
}

func (c *Client) M() *Request { return &Request{client: c, use: MemoryClient} }
func (c *Client) R() *Request { return &Request{client: c, use: RedisClient} }
