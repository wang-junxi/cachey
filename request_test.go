package cachey

import (
	"fmt"
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
	r.SetResultType(&result)
	if reflect.ValueOf(r.result).Elem().Type() != reflect.TypeOf(result) {
		t.Errorf("Unexpected type of result: got %T, expected %T", reflect.ValueOf(r.result).Elem().Type(), reflect.TypeOf(result))
	}

	// when type of result is 'int'
	var result2 int
	r.SetResultType(&result2)
	if reflect.ValueOf(r.result).Elem().Type() != reflect.TypeOf(result2) {
		t.Errorf("Unexpected type of result: got %T, expected %T", reflect.ValueOf(r.result).Elem().Type(), reflect.TypeOf(result2))
	}

	// when type of result is 'Person'
	type Person struct {
		Name string
		Age  int
	}
	var result3 Person
	r.SetResultType(&result3)
	if reflect.ValueOf(r.result).Elem().Type() != reflect.TypeOf(result3) {
		t.Errorf("Unexpected type of result: got %T, expected %T", reflect.ValueOf(r.result).Elem().Type(), reflect.TypeOf(result3))
	}
}

func TestRequest_Execute(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	var (
		_loopNum  = 5
		_cacheKey = "fake-cache-key"
		_expire   = time.Minute
	)

	newClient := func() *Client {
		fakeRedis := miniredis.RunT(t)
		rc := redis.NewClient(&redis.Options{Addr: fakeRedis.Addr()})
		mc := cache.New(1*time.Minute, 10*time.Minute)
		return New(rc, mc).EnableDebug()
	}

	t.Run("when CacheKey is not set", func(t *testing.T) {
		f := func(args ...interface{}) (interface{}, error) { return 25, nil }
		_, err := newClient().M().SetFunc(f).SetResultType(0).Execute()
		if err != nil {
			t.Errorf("Execute() error = %v, wantErr %v", err, nil)
		}
	})

	defaultGotAssert := func(got interface{}) bool { return true }

	tests := []struct {
		name      string
		f         Func
		result    interface{}
		gotAssert func(got interface{}) bool
		wantErr   bool
	}{
		{
			name:      "when Func is not set",
			f:         nil,
			result:    0,
			gotAssert: defaultGotAssert,
			wantErr:   true, // you must set Func and ResultType, wantErr false
		},
		{
			name:      "when ResultType is not set",
			f:         func(args ...interface{}) (interface{}, error) { return 25, nil },
			result:    nil,
			gotAssert: defaultGotAssert,
			wantErr:   true, // you must set Func and ResultType, wantErr false
		},
		{
			name:      "when Func returns error",
			f:         func(args ...interface{}) (interface{}, error) { return nil, fmt.Errorf("fake-error") },
			result:    0,
			gotAssert: defaultGotAssert,
			wantErr:   true,
		},
		{
			name:      "when result is 'int' value",
			f:         func(args ...interface{}) (interface{}, error) { return 25, nil },
			result:    0,
			gotAssert: func(got interface{}) bool { return got.(int) == 25 },
			wantErr:   false,
		},
		{
			name:      "when result is 'uint' value",
			f:         func(args ...interface{}) (interface{}, error) { return uint(25), nil },
			result:    uint(0),
			gotAssert: func(got interface{}) bool { return got.(uint) == 25 },
			wantErr:   false,
		},
		{
			name:      "when result is 'float64' value",
			f:         func(args ...interface{}) (interface{}, error) { return 1.23, nil },
			result:    0.0,
			gotAssert: func(got interface{}) bool { return got.(float64) == 1.23 },
			wantErr:   false,
		},
		{
			name:      "when result is 'string' value",
			f:         func(args ...interface{}) (interface{}, error) { return "fake-str", nil },
			result:    "",
			gotAssert: func(got interface{}) bool { return got.(string) == "fake-str" },
			wantErr:   false,
		},
		{
			name:      "when result is 'bool' value",
			f:         func(args ...interface{}) (interface{}, error) { return true, nil },
			result:    false,
			gotAssert: func(got interface{}) bool { return got.(bool) == true },
			wantErr:   false,
		},
		{
			name:   "when result is '[]int' value",
			f:      func(args ...interface{}) (interface{}, error) { return []int{1, 2, 3}, nil },
			result: []int{},
			gotAssert: func(got interface{}) bool {
				return got.([]int)[0] == 1 && got.([]int)[1] == 2 && got.([]int)[2] == 3
			},
			wantErr: false,
		},
		{
			name:   "when result is 'map[string]int' value",
			f:      func(args ...interface{}) (interface{}, error) { return map[string]int{"fake-name": 25}, nil },
			result: map[string]int{},
			gotAssert: func(got interface{}) bool {
				return got.(map[string]int)["fake-name"] == 25
			},
			wantErr: false,
		},
		{
			name:   "when result is 'Person' value",
			f:      func(args ...interface{}) (interface{}, error) { return Person{Name: "fake-name", Age: 25}, nil },
			result: Person{},
			gotAssert: func(got interface{}) bool {
				return got.(Person).Name == "fake-name" && got.(Person).Age == 25
			},
			wantErr: false,
		},
		{
			name: "when result is '[]Person' value",
			f: func(args ...interface{}) (interface{}, error) {
				return []Person{{Name: "fake-name", Age: 25}, {Name: "fake-name", Age: 25}}, nil
			},
			result: []Person{},
			gotAssert: func(got interface{}) bool {
				return got.([]Person)[0].Name == "fake-name" && got.([]Person)[1].Age == 25
			},
			wantErr: false,
		},
		{
			name: "when result is 'map[string]Person' value",
			f: func(args ...interface{}) (interface{}, error) {
				return map[string]Person{"person1": {Name: "fake-name", Age: 25}, "person2": {Name: "fake-name", Age: 25}}, nil
			},
			result: map[string]Person{},
			gotAssert: func(got interface{}) bool {
				return got.(map[string]Person)["person1"].Name == "fake-name" && got.(map[string]Person)["person2"].Age == 25
			},
			wantErr: false,
		},
		{
			name:   "when result is '*Person' value",
			f:      func(args ...interface{}) (interface{}, error) { return &Person{Name: "fake-name", Age: 25}, nil },
			result: &Person{},
			gotAssert: func(got interface{}) bool {
				return got.(*Person).Name == "fake-name" && got.(*Person).Age == 25
			},
			wantErr: false,
		},
		{
			name: "when result is '[]*Person' value",
			f: func(args ...interface{}) (interface{}, error) {
				return []*Person{{Name: "fake-name", Age: 25}, {Name: "fake-name", Age: 25}}, nil
			},
			result: []*Person{},
			gotAssert: func(got interface{}) bool {
				return got.([]*Person)[0].Name == "fake-name" && got.([]*Person)[1].Age == 25
			},
			wantErr: false,
		},
		{
			name: "when result is 'map[string]*Person' value",
			f: func(args ...interface{}) (interface{}, error) {
				return map[string]*Person{"person1": {Name: "fake-name", Age: 25}, "person2": {Name: "fake-name", Age: 25}}, nil
			},
			result: map[string]*Person{},
			gotAssert: func(got interface{}) bool {
				return got.(map[string]*Person)["person1"].Name == "fake-name" && got.(map[string]*Person)["person2"].Age == 25
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newClient()
			for _, req := range []*Request{c.M(), c.R()} {
				for i := 0; i < _loopNum; i++ {
					got, err := req.SetCacheKey(_cacheKey).SetExpiration(_expire).SetFunc(tt.f).SetResultType(tt.result).Execute()
					if (err != nil) != tt.wantErr {
						t.Errorf("Request.Execute() error = %v, wantErr %v", err, tt.wantErr)
						return
					}
					if !tt.gotAssert(got) {
						t.Errorf("Request.Execute() got is not expected, got %v", got)
					}
				}
			}
		})
	}
}
