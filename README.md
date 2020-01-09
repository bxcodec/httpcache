# httpcache, inject-able HTTP cache in Golang

### Based on RFCC 7234

Howdy there!!!

Usually when we want to integrate with cache (let's say Redis), we usually have to do many changes in our code. 
What if, we just inject the cache to the HTTP client. So we don't have to create many changes in every line of our code to support the cache features?
With only less than 10 line of code, you can got a complete implementations of HTTP Cache based on [RFC 7234](http://tools.ietf.org/html/rfc7234)

[![Build Status](https://travis-ci.com/bxcodec/httpcache.svg?token=Y64SjWyDK7wXJiFFqV6M&branch=master)](https://travis-ci.com/bxcodec/httpcache)
[![License](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/bxcodec/httpcache/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/bxcodec/httpcache?status.svg)](https://godoc.org/github.com/bxcodec/httpcache)

This package is used for caching your http request results from the server. Example how to use can be seen below.

## Index

* [Support](#support)
* [Getting Started](#getting-started)
* [Example](#example) 
* [Contribution](#contribution)


## Support

You can file an [Issue](https://github.com/bxcodec/httpcache/issues/new).
See documentation in [Godoc](https://godoc.org/github.com/bxcodec/httpcache)


## Getting Started

#### Download

```shell
go get -u github.com/bxcodec/httpcache
```
# Example

---

Example how to use more details can be seen in the sample folder: [/sample](/sample)

Short example:

```go

// Inject the HTTP Client with httpcache
client := &http.Client{}
err := httpcache.NewWithInmemoryCache(client, time.Second*60)
if err != nil {
  log.Fatal(err)
}
 
// And your HTTP Client already supported for HTTP Cache
// To verify you can run a request in a loop

for i:=0; i< 10; i++ {
  startTime := time.Now()
  req, err := http.NewRequest("GET", "https://bxcodec.io", nil)
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

### TODOs
- [ ] Add Redis Storage


## Inspirations and Thanks
- [pquerna/cachecontrol](https://github.com/pquerna/cachecontrol) for the Cache-Header Extraction
- [bxcodec/gothca](https://github.com/bxcodec/gotcha) for in-memory cache. _*Notes: if you find another library that has a better way for inmemm cache, please raise an issue and submit a PR_


## Contribution
---

To contrib to this project, you can open a PR or an issue.

