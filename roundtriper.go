package hache

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// RoundTrip custom plugable' struct of implementation of the http.RoundTripper
type RoundTrip struct {
	DefaultRoundTripper http.RoundTripper
	CacheInteractor     CacheInteractor
}

// RoundTrip the implementation of http.RoundTripper
func (r *RoundTrip) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if !isTheHeaderAllowCachedResponse(req) {
		return r.DefaultRoundTripper.RoundTrip(req)
	}

	if !isHTTPMethodValid(req) {
		return r.DefaultRoundTripper.RoundTrip(req)
	}

	resp, err = getCachedResponse(r.CacheInteractor, req)
	if resp != nil && err == nil {
		buildTheCachedResponseHeader(resp)
		return
	}

	resp, err = r.DefaultRoundTripper.RoundTrip(req)
	if err != nil {
		return
	}
	storeRespToCache(r.CacheInteractor, req, resp)
	return
}

func storeRespToCache(cacheInteractor CacheInteractor, req *http.Request, resp *http.Response) (err error) {
	cachedResp := CachedResponse{
		StatusCode:    resp.StatusCode,
		RequestMethod: req.Method,
		RequestURI:    req.RequestURI,
		CachedTime:    time.Now(),
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	cachedResp.DumpedResponse = bodyBytes
	err = cacheInteractor.Set(getCacheKey(req), cachedResp)
	return
}

func getCachedResponse(cacheInteractor CacheInteractor, req *http.Request) (resp *http.Response, err error) {
	item, err := cacheInteractor.Get(getCacheKey(req))
	if err != nil {
		return
	}

	cachedResp, ok := item.(CachedResponse)
	if !ok {
		return
	}
	if err = cachedResp.Validate(); err != nil {
		return
	}

	cachedResponse := bytes.NewBuffer(cachedResp.DumpedResponse)
	resp, err = http.ReadResponse(bufio.NewReader(cachedResponse), req)
	if err != nil {
		return
	}
	resp.StatusCode = cachedResp.StatusCode
	return
}

func getCacheKey(req *http.Request) (key string) {
	key = fmt.Sprintf("%s %s", req.Method, req.RequestURI)
	return
}

// buildTheCachedResponse will finalize the response header
func buildTheCachedResponseHeader(resp *http.Response) {
	panic("TODO: (bxcodec) Add the header based on RFC 7234")
}

// check the header if the response will cached or not
func isTheHeaderAllowCachedResponse(req *http.Request) bool {
	panic("TODO: (bxcodec) check the header based on RFC 7234")
}

func isHTTPMethodValid(req *http.Request) bool {
	panic("TODO: (bxcodec) check the method verb based on RFC 7234")
}
