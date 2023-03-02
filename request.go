package cachey

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"
)

type ClientType uint8

const (
	RedisClient ClientType = iota
	MemoryClient
)

type Func func(args ...interface{}) (interface{}, error)

type Request struct {
	client *Client
	use    ClientType

	cacheKey   string
	expiration time.Duration

	f      Func
	result interface{}
}

func (r *Request) SetCacheKey(cacheKey string) *Request {
	r.cacheKey = cacheKey
	return r
}

func (r *Request) SetExpiration(expiration time.Duration) *Request {
	r.expiration = expiration
	return r
}

func (r *Request) SetFunc(f Func) *Request {
	r.f = f
	return r
}

func (r *Request) SetResult(result interface{}) *Request {
	r.result = getPointer(result)
	return r
}

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

func (r *Request) get() error {
	switch r.use {
	case MemoryClient:
		if val, hit := r.client.memoryClient.Get(r.cacheKey); hit {
			r.result = val
			return nil
		}
		return fmt.Errorf("cacheKey '%s' not hit with memoryClient", r.cacheKey)

	case RedisClient:
		bytes, err := r.client.redisClient.Get(context.Background(), r.cacheKey).Bytes()
		if err != nil {
			return fmt.Errorf("get cacheKey '%s' failed with redisClient, reason: %s", r.cacheKey, err.Error())
		}
		return json.Unmarshal(bytes, r.result)

	default:
		return fmt.Errorf("method 'get' is not implemented for Enum_ClientType: %d", r.use)
	}
}

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

func getPointer(v interface{}) interface{} {
	vv := valueOf(v)
	if vv.Kind() == reflect.Ptr {
		return v
	}
	return reflect.New(vv.Type()).Interface()
}

func valueOf(i interface{}) reflect.Value {
	return reflect.ValueOf(i)
}
