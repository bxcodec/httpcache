# httpcache, inject-able HTTP cache in Golang

**Howdy there!!!**

Usually when we want to integrate with cache (let's say Redis), we usually have to do many changes in our code.
What if, we just inject the cache to the HTTP client. So we don't have to create many changes in every line of our code to support the cache features?
With only less than 10 line of code, you can got a complete implementations of HTTP Cache based on [RFC 7234](http://tools.ietf.org/html/rfc7234)

[![Build Status](https://travis-ci.com/bxcodec/httpcache.svg?token=Y64SjWyDK7wXJiFFqV6M&branch=master)](https://travis-ci.com/bxcodec/httpcache)
[![License](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/bxcodec/httpcache/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/bxcodec/httpcache?status.svg)](https://godoc.org/github.com/bxcodec/httpcache)
[![Go.Dev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/bxcodec/httpcache?tab=doc)

This package is used for caching your http request results from the server. Example how to use can be seen below.

## Index

- [Support](#support)
- [Getting Started](#getting-started)
- [Example](#example)
- [Contribution](#contribution)

## Support

You can file an [Issue](https://github.com/bxcodec/httpcache/issues/new).
See documentation in [Godoc](https://godoc.org/github.com/bxcodec/httpcache) or in [go.dev](https://pkg.go.dev/github.com/bxcodec/httpcache?tab=doc)

## Getting Started

#### Download

```shell
go get -u github.com/bxcodec/httpcache
```

# Example with Inmemory Storage

---

Example how to use more details can be seen in the example file: [./example_inmemory_storage_test.go](./example_inmemory_storage_test.go)

Short example:

```go

// Inject the HTTP Client with httpcache
client := &http.Client{}
_, err := httpcache.NewWithInmemoryCache(client, true, time.Second*60)
if err != nil {
  log.Fatal(err)
}

// And your HTTP Client already supported for HTTP Cache
// To verify you can run a request in a loop

for i:=0; i< 10; i++ {
  startTime := time.Now()
  req, err := http.NewRequest("GET", "https://imantumorang.com", nil)
  if err != nil {
    log.Fatal((err))
  }

  res, err := client.Do(req)
  if err != nil {
    log.Fatal(err)
  }

  fmt.Printf("Response time: %vms\n", time.Since(startTime).Microseconds())
  fmt.Println("Status Code", res.StatusCode)
}
// See the response time, it will different on each request and will go smaller.
```

# Example with Custom Storage

You also can use your own custom storage, what you need to do is implement the `cache.ICacheInteractor` interface.
Example how to use more details can be seen in the example file: [./example_custom_storage_test.go](./example_custom_storage_test.go)

Example:

```go
client := &http.Client{}
_, err := httpcache.NewWithCustomStorageCache(client,true, mystorage.NewCustomInMemStorage())
if err != nil {
	log.Fatal(err)
}
```

### About RFC 7234 Compliance

You can disable/enable the RFC Compliance as you want. If RFC 7234 is too complex for you, you can just disable it by set the RFCCompliance parameter to false

```go
_, err := httpcache.NewWithInmemoryCache(client, false, time.Second*60)
// or
_, err := httpcache.NewWithCustomStorageCache(client,false, mystorage.NewCustomInMemStorage())
```

The downside of disabling the RFC Compliance, **All the response/request will be cached automatically**. Do with caution.

### TODOs

- See the [issues](https://github.com/bxcodec/httpcache/issues)

## Inspirations and Thanks

- [pquerna/cachecontrol](https://github.com/pquerna/cachecontrol) for the Cache-Header Extraction
- [bxcodec/gothca](https://github.com/bxcodec/gotcha) for in-memory cache. _\*Notes: if you find another library that has a better way for inmemm cache, please raise an issue or submit a PR_

## Contribution

---

To contrib to this project, you can open a PR or an issue.
