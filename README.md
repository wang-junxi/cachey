<div align="center" id="top"> 
  <img src="./.github/cachey.png" alt="Cachey" />
  &#xa0;
</div>

<h1 align="center">Cachey</h1>

<p align="center">
  <img alt="Github top language" src="https://img.shields.io/github/languages/top/wang-junxi/cachey?color=56BEB8">

  <img alt="Github language count" src="https://img.shields.io/github/languages/count/wang-junxi/cachey?color=56BEB8">

  <img alt="Repository size" src="https://img.shields.io/github/repo-size/wang-junxi/cachey?color=56BEB8">

  <img alt="License" src="https://img.shields.io/github/license/wang-junxi/cachey?color=56BEB8">
</p>


<p align="center">
  <a href="#dart-about">About</a> &#xa0; | &#xa0; 
  <a href="#sparkles-features">Features</a> &#xa0; | &#xa0;
  <a href="#rocket-technologies">Technologies</a> &#xa0; | &#xa0;
  <a href="#white_check_mark-requirements">Requirements</a> &#xa0; | &#xa0;
  <a href="#checkered_flag-starting">Starting</a> &#xa0; | &#xa0;
  <a href="#memo-license">License</a> &#xa0; | &#xa0;
  <a href="https://github.com/wang-junxi" target="_blank">Author</a>
</p>

<br>

## :dart: About ##

Cachey is a simple, easy-to-use caching of function values based on redis or memory in Go.

## :sparkles: Features ##

:heavy_check_mark: Simple and chainable methods for settings and execute;\
:heavy_check_mark: Predefined result structure to handle function return value;\
:heavy_check_mark: Auto unmarshal result;

## :rocket: Technologies ##

The following tools were used in this project:

- [go-redis](https://github.com/redis/go-redis)
- [go-cache](https://github.com/patrickmn/go-cache)
- [zerolog](https://github.com/rs/zerolog)

## :white_check_mark: Requirements ##

Before starting :checkered_flag:, you need to have [Git](https://git-scm.com) and [Go](https://go.dev/doc/install) installed.

## :checkered_flag: Install ##

```bash
go get -u github.com/wang-junxi/cachey
```

## :checkered_flag: Use Example ##

Using memory to cache function values 
```golang
mc := cache.New(time.Hour, time.Hour)
c := New(nil, mc)
// c := New(nil, nil) // Or just use memory cache with default config

// when caching 'string' value with memory
var (
  strPlaceholder string
  getName        = func(args ...interface{}) (interface{}, error) {
    return "fake-name", nil
  }
)

res, err := c.M().
  SetCacheKey("cache_key_name").
  SetFunc(getName).
  SetResult(strPlaceholder).
  Execute()

fmt.Println(res.(string), err)

// when caching 'Person' struct with memory
type Person struct {
  Name string
  Age  int
}

var (
  person    = Person{Name: "fake-name", Age: 25}
  getPerson = func(args ...interface{}) (interface{}, error) {
    return person, nil
  }
)

res, err = c.M().
  SetCacheKey("cache_key_person").
  SetFunc(getPerson).
  SetResult(Person{}).
  Execute()

fmt.Println(res.(Person), err)
```

Using redis to cache function values 
```golang
rc := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
c := New(rc, nil)

// when caching '[]int' slice with redis
var (
  intSlicePlaceholder []int
  getAges             = func(args ...interface{}) (interface{}, error) {
    return []int{25, 21, 28}, nil
  }
)

res, err := c.R().
  SetCacheKey("cache_key_ages").
  SetFunc(getAges).
  SetResult(intSlicePlaceholder).
  Execute()

fmt.Println(res.([]int), err)

// when caching '[]Person' slice with redis
type Person struct {
  Name string
  Age  int
}

var (
  person     = &Person{Name: "fake-name", Age: 25}
  persons    = []*Person{person, person}
  getPersons = func(args ...interface{}) (interface{}, error) {
    return persons, nil
  }
)

res, err = c.R().
  SetCacheKey("cache_key_persons").
  SetFunc(getPersons).
  SetResult([]*Person{}).
  Execute()

fmt.Println(res.([]*Person), err)
```

## :memo: License ##

This project is under license from MIT. For more details, see the [LICENSE](LICENSE.md) file.

&#xa0;

<a href="#top">Back to top</a>
