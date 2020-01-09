# Docs

Howdy there!!!

Usually when we want to integrate with cache (let's say Redis), we usually have to do many changes in our code. 
What if, we just inject the cache to the HTTP client. So we don't have to create many changes in each line of our code to get the data from Cache, do the validation etc.

## Introduce Hache: Injecte-able HTTP Cache for Golang HTTP Client

[![Build Status](https://travis-ci.com/bxcodec/hache.svg?token=Y64SjWyDK7wXJiFFqV6M&branch=master)](https://travis-ci.org/bxcodec/hache)
[![License](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/bxcodec/hache/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/bxcodec/hache?status.svg)](https://godoc.org/github.com/bxcodec/hache)

This package is used for caching your http request results from the server. Example how to use can be seen below.

## Index

* [Support](#support)
* [Getting Started](#getting-started)
* [Example](#example)
* [Limitation](#limitation)
* [Contribution](#contribution)


## Support

You can file an [Issue](https://github.com/bxcodec/hache/issues/new).
See documentation in [Godoc](https://godoc.org/github.com/bxcodec/hache)


## Getting Started

#### Download

```shell
go get -u github.com/bxcodec/hache
```
# Example

---

Example how to use more details can be seen in the sample folder: [/sample](/sample)

Short example:

```go

// Inject the HTTP Client with Hache
client := &http.Client{}
err := hache.NewWithInmemoryCache(client, time.Second*60)
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

### Inject with your Redis Service
//TODO(bxcodec)



## Inspirations and Thanks
- [pquerna/cachecontrol](https://github.com/pquerna/cachecontrol) for the Cache-Header Extraction


## Contribution
---

To contrib to this project, you can open a PR or an issue.

