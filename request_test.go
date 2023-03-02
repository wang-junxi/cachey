package cachey

import (
	"reflect"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/patrickmn/go-cache"
)

func TestRequest_validate(t *testing.T) {
	// test when use == MemoryClient and memoryClient is nil
	r := &Request{
		client:   &Client{},
		use:      MemoryClient,
		cacheKey: "test_key",
	}
	err := r.validate()
	if err != nil {
		t.Errorf("validate failed, expected: nil, got: %s", err.Error())
	}

	// test when use == RedisClient and redisClient is nil
	r = &Request{
		client:   &Client{},
		use:      RedisClient,
		cacheKey: "test_key",
	}
	err = r.validate()
	if err == nil {
		t.Errorf("validate failed, expected: %s, got: nil", "redis client is not initialized")
	}

	// test when cacheKey is not set
	r = &Request{
		client: &Client{
			memoryClient: cache.New(time.Hour, time.Hour),
		},
		use: MemoryClient,
	}
	err = r.validate()
	if err == nil {
		t.Errorf("validate failed, expected: %s, got: nil", "cacheKey is not set")
	}
}

func TestRequest_SetResult(t *testing.T) {
	r := &Request{}

	// when type of result is 'string'
	var result string
	r.SetResult(&result)
	if reflect.ValueOf(r.result).Elem().Type() != reflect.TypeOf(result) {
		t.Errorf("Unexpected type of result: got %T, expected %T", reflect.ValueOf(r.result).Elem().Type(), reflect.TypeOf(result))
	}

	// when type of result is 'int'
	var result2 int
	r.SetResult(&result2)
	if reflect.ValueOf(r.result).Elem().Type() != reflect.TypeOf(result2) {
		t.Errorf("Unexpected type of result: got %T, expected %T", reflect.ValueOf(r.result).Elem().Type(), reflect.TypeOf(result2))
	}

	// when type of result is 'Person'
	type Person struct {
		Name string
		Age  int
	}
	var result3 Person
	r.SetResult(&result3)
	if reflect.ValueOf(r.result).Elem().Type() != reflect.TypeOf(result3) {
		t.Errorf("Unexpected type of result: got %T, expected %T", reflect.ValueOf(r.result).Elem().Type(), reflect.TypeOf(result3))
	}
}

func TestRequest_get_set(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	var (
		_loopNum = 5
		_name    = "fake-name"
		_age     = 25
		_ages    = []int{25, 21, 28}
		_person  = Person{Name: _name, Age: _age}
		_persons = []Person{_person, _person}
	)

	// create Client
	fakeRedis := miniredis.RunT(t)
	rc := redis.NewClient(&redis.Options{Addr: fakeRedis.Addr()})
	mc := cache.New(1*time.Minute, 10*time.Minute)
	c := New(rc, mc)

	t.Run("when caching 'struct Person' value", func(t *testing.T) {
		var (
			_cacheKeyPerson = "test_cache_key_person"
			_getPerson      = func(args ...interface{}) (interface{}, error) {
				time.Sleep(time.Second * 1) // mock processing of the function
				return _person, nil
			}
		)

		for _, req := range []*Request{c.M(), c.R()} {
			for i := 0; i < _loopNum; i++ {
				res, err := req.SetCacheKey(_cacheKeyPerson).SetFunc(_getPerson).SetResult(Person{}).Execute()
				if res.(Person).Name != _name || res.(Person).Age != _age {
					t.Errorf("Unexpected value of result: got %v, expected %v", res, _person)
				}
				if err != nil {
					t.Errorf("Unexpected err: got %v, expected nil", err)
				}
			}
		}
	})

	t.Run("when caching 'string' value", func(t *testing.T) {
		var (
			_strPlaceholder string
			_cacheKeyName   = "test_cache_key_name"
			_getName        = func(args ...interface{}) (interface{}, error) {
				time.Sleep(time.Second * 1) // mock processing of the function
				return _name, nil
			}
		)

		for _, req := range []*Request{c.M(), c.R()} {
			for i := 0; i < _loopNum; i++ {
				res, err := req.SetCacheKey(_cacheKeyName).SetFunc(_getName).SetResult(_strPlaceholder).Execute()
				if res.(string) != _name {
					t.Errorf("Unexpected value of result: got %v, expected %v", res, _name)
				}
				if err != nil {
					t.Errorf("Unexpected err: got %v, expected nil", err)
				}
			}
		}
	})

	t.Run("when caching '[]int' value", func(t *testing.T) {
		var (
			_slicePlaceholder []int
			_cacheKeyAges     = "test_cache_key_ags"
			_getAges          = func(args ...interface{}) (interface{}, error) {
				time.Sleep(time.Second * 1) // mock processing of the function
				return _ages, nil
			}
		)

		for _, req := range []*Request{c.M(), c.R()} {
			for i := 0; i < _loopNum; i++ {
				res, err := req.SetCacheKey(_cacheKeyAges).SetFunc(_getAges).SetResult(_slicePlaceholder).Execute()
				if res.([]int)[0] != _ages[0] || res.([]int)[1] != _ages[1] || res.([]int)[2] != _ages[2] {
					t.Errorf("Unexpected value of result: got %v, expected %v", res, _ages)
				}
				if err != nil {
					t.Errorf("Unexpected err: got %v, expected nil", err)
				}
			}
		}
	})

	t.Run("when caching '[]Person' value", func(t *testing.T) {
		var (
			_cacheKeyPersons = "test_cache_key_persons"
			_getPersons      = func(args ...interface{}) (interface{}, error) {
				time.Sleep(time.Second * 1) // mock processing of the function
				return _persons, nil
			}
		)

		for _, req := range []*Request{c.M(), c.R()} {
			for i := 0; i < _loopNum; i++ {
				res, err := req.SetCacheKey(_cacheKeyPersons).SetFunc(_getPersons).SetResult([]Person{}).Execute()
				if res.([]Person)[0] != _person || res.([]Person)[1] != _person {
					t.Errorf("Unexpected value of result: got %v, expected %v", res, _persons)
				}
				if err != nil {
					t.Errorf("Unexpected err: got %v, expected nil", err)
				}
			}
		}
	})
}
