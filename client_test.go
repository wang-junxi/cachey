package cachey

import (
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog"
)

func TestClient(t *testing.T) {
	rc := &redis.Client{}
	mc := &cache.Cache{}

	// test function New()
	c := New(rc, mc)
	if c == nil {
		t.Error("Expected non-nil Client")
	}

	// test method client.M()
	r := c.M()
	if r.use != MemoryClient {
		t.Errorf("Expected client type MemoryClient; got %d", r.use)
	}

	// test method client.R()
	r = c.R()
	if r.use != RedisClient {
		t.Errorf("Expected client type RedisClient; got %d", r.use)
	}

	// test method client.EnableDebug
	c = c.EnableDebug()
	if c.logger.GetLevel() != zerolog.DebugLevel {
		t.Errorf("Expected log level 'Warn'; got %d", c.logger.GetLevel())
	}
}
