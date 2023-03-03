// Package cachey provides a simple and flexible way to cache the results of functions
// using either redis or memory as the backend.
package cachey

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
)

// ClientType represents the type of cache backend to use for a request.
type ClientType uint8

const (
	RedisClient  ClientType = iota // Use redis as the cache backend
	MemoryClient                   // Use memory as the cache backend
)

// Func represents a function whose result can be cached by cachey.
type Func func(args ...interface{}) (interface{}, error)

// Request represents a cachey request that can execute a function and cache its result
// using either redis or memory as the cache backend.
type Request struct {
	client *Client
	use    ClientType

	cacheKey   string
	expiration time.Duration

	f      Func
	result interface{}
}

// SetCacheKey sets the cache key for this request and returns itself for chaining.
func (r *Request) SetCacheKey(cacheKey string) *Request {
	r.cacheKey = cacheKey
	return r
}

// SetExpiration sets the expiration time for caching the result of this request and returns itself for chaining.
func (r *Request) SetExpiration(expiration time.Duration) *Request {
	r.expiration = expiration
	return r
}

// SetFunc sets the function to execute and cache its result for this request and returns itself for chaining.
func (r *Request) SetFunc(f Func) *Request {
	r.f = f
	return r
}

// SetResultType sets the type of result expected from the function execution for this request and returns itself for chaining.
func (r *Request) SetResultType(result interface{}) *Request {
	r.result = result
	return r
}

// validate validates if this request is valid and ready to execute. It returns an error if any validation fails.
func (r *Request) validate() error {
	if r.use == MemoryClient && r.client.memoryClient == nil {
		r.client.memoryClient = cache.New(time.Hour, time.Hour)
	}

	if r.use == RedisClient && r.client.redisClient == nil {
		return fmt.Errorf("redis client is not initialized")
	}

	if r.cacheKey == "" {
		return fmt.Errorf("cacheKey is not set")
	}

	return nil
}

// Execute executes this request by first trying to get the result from the cache with the given key. If it fails, it executes
// the function and caches its result with the given key and expiration time. It returns the result and an error if any.
func (r *Request) Execute(args ...interface{}) (interface{}, error) {
	// validate members of object Request
	if r.f == nil || r.result == nil {
		return r.result, fmt.Errorf("Func or result is not set")
	}

	if err := r.validate(); err != nil {
		log.Warn().Msgf("cachey is not in effect. reason:%s.", err.Error())
	}

	// try to get val from cache with cacheKey
	if err := r.get(); err != nil {
		log.Warn().Msg(err.Error())
	} else {
		return r.result, nil
	}

	// execute the function
	if res, err := r.f(args); err != nil {
		return r.result, fmt.Errorf("Execute Func failed, reason: %s", err.Error())
	} else {
		r.result = res
	}

	// try to set val to cache with cacheKey
	if err := r.set(); err != nil {
		log.Warn().Msg(err.Error())
	}
	return r.result, nil
}

// get tries to get the result from the cache with the given key and unmarshal it into the expected result type. It returns an error if it fails.
func (r *Request) get() error {
	switch r.use {
	case MemoryClient:
		if val, hit := r.client.memoryClient.Get(r.cacheKey); hit {
			r.result = val
			return nil
		}
		return fmt.Errorf("cacheKey '%s' not hit with memoryClient", r.cacheKey)

	case RedisClient:
		data, err := r.client.redisClient.Get(context.Background(), r.cacheKey).Bytes()
		if err != nil {
			return fmt.Errorf("get cacheKey '%s' failed with redisClient, reason: %s", r.cacheKey, err.Error())
		}
		return r.unmarshal(data)

	default:
		return fmt.Errorf("method 'get' is not implemented for Enum_ClientType: %d", r.use)
	}
}

// set tries to marshal the result into a byte slice and set it to the cache with the given key and expiration time. It returns an error if it fails.
func (r *Request) set() error {
	switch r.use {
	case MemoryClient:
		r.client.memoryClient.Set(r.cacheKey, r.result, r.expiration)
		return nil

	case RedisClient:
		if bytes, err := json.Marshal(r.result); err != nil {
			return fmt.Errorf("marshal result (%v) to cacheKey '%s' failed with redisClient, reason: %s", r.result, r.cacheKey, err.Error())
		} else {
			return r.client.redisClient.Set(context.Background(), r.cacheKey, string(bytes), r.expiration).Err()
		}

	default:
		return fmt.Errorf("method 'set' is not implemented for Enum_ClientType: %d", r.use)
	}
}

// unmarshal tries to unmarshal value into result type using mapstructure or json depending on result type kind. It returns an error if it fails.
func (r *Request) unmarshal(data []byte) error {
	kind := reflect.ValueOf(r.result).Kind()
	switch kind {
	case reflect.Pointer:
		return json.Unmarshal(data, r.result)

	case reflect.Slice:
		var meta []interface{}
		if err := json.Unmarshal(data, &meta); err != nil {
			return err
		}
		return mapstructure.WeakDecode(meta, &r.result)

	case reflect.Struct:
		meta := new(map[string]interface{})
		if err := json.Unmarshal(data, meta); err != nil {
			return err
		}
		return mapstructure.WeakDecode(meta, &r.result)

	default:
		var meta interface{}
		if err := json.Unmarshal(data, &meta); err != nil {
			return err
		}
		return mapstructure.WeakDecode(meta, &r.result)
	}
}
