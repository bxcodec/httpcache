package hache

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/bxcodec/hache/cache"
)

// Headers
const (
	HeaderAuthorization = "Authorization"
	HeaderCacheControl  = "Cache-Control"
)

var (
	// CacheAuthorizedRequest used for determine that a request with Authorization header should be cached or not
	CacheAuthorizedRequest = false // TODO(bxcodec): Need to revised about this feature
)

// RoundTrip custom plugable' struct of implementation of the http.RoundTripper
type RoundTrip struct {
	DefaultRoundTripper http.RoundTripper
	CacheInteractor     cache.Interactor
}

// NewRoundtrip will create an implementations of cache http roundtripper
func NewRoundtrip(defaultRoundTripper http.RoundTripper, cacheActor cache.Interactor) http.RoundTripper {
	return &RoundTrip{
		DefaultRoundTripper: defaultRoundTripper,
		CacheInteractor:     cacheActor,
	}
}

// RoundTrip the implementation of http.RoundTripper
func (r *RoundTrip) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if allowedFromCache(req) {
		resp, cachedItem, err := getCachedResponse(r.CacheInteractor, req)
		if resp != nil && err == nil {
			buildTheCachedResponseHeader(resp, cachedItem)
			return resp, err
		}
	}
	err = nil
	resp, err = r.DefaultRoundTripper.RoundTrip(req)
	if err != nil {
		return
	}

	if !allowedToCache(req, resp) {
		return
	}

	err = storeRespToCache(r.CacheInteractor, req, resp)
	if err != nil {
		log.Println(err)
	}

	return
}

func storeRespToCache(cacheInteractor cache.Interactor, req *http.Request, resp *http.Response) (err error) {
	cachedResp := cache.CachedResponse{
		RequestMethod: req.Method,
		RequestURI:    req.RequestURI,
		CachedTime:    time.Now(),
	}

	dumpedResponse, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return
	}
	cachedResp.DumpedResponse = dumpedResponse
	err = cacheInteractor.Set(getCacheKey(req), cachedResp)
	return
}

func getCachedResponse(cacheInteractor cache.Interactor, req *http.Request) (resp *http.Response, cachedResp cache.CachedResponse, err error) {
	cachedResp, err = cacheInteractor.Get(getCacheKey(req))
	if err != nil {
		return
	}

	cachedResponse := bytes.NewBuffer(cachedResp.DumpedResponse)
	resp, err = http.ReadResponse(bufio.NewReader(cachedResponse), req)
	if err != nil {
		return
	}

	return
}

func getCacheKey(req *http.Request) (key string) {
	key = fmt.Sprintf("%s %s", req.Method, req.RequestURI)
	if (CacheAuthorizedRequest ||
		(strings.ToLower(req.Header.Get(HeaderCacheControl)) == "private")) &&
		req.Header.Get(HeaderAuthorization) != "" {
		key = fmt.Sprintf("%s %s", key, req.Header.Get(HeaderAuthorization))
	}
	return
}

// buildTheCachedResponse will finalize the response header
func buildTheCachedResponseHeader(resp *http.Response, cachedResp cache.CachedResponse) {
	resp.Header.Add("Expires", cachedResp.CachedTime.String())
	// TODO: (bxcodec) add more headers related to cache
}

// check the header if the response will cached or not
func allowedToCache(req *http.Request, resp *http.Response) (ok bool) {
	// A request with authorization header must not be cached
	// https://tools.ietf.org/html/rfc7234#section-3.2
	// Unless configured by user to cache request by authorization
	if ok = (!CacheAuthorizedRequest && req.Header.Get(HeaderAuthorization) == ""); !ok {
		return
	}

	// check if the request method allowed to be cached
	if ok = requestMethodValid(req); !ok {
		return
	}

	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control#Preventing_caching
	if ok = strings.ToLower(req.Header.Get(HeaderCacheControl)) != "no-store"; !ok {
		return
	}
	if ok = strings.ToLower(resp.Header.Get(HeaderCacheControl)) != "no-store"; !ok {
		return
	}

	// Only cache the response of with code 200
	if ok = resp.StatusCode == http.StatusOK; !ok {
		return
	}
	return
}

func allowedFromCache(req *http.Request) (ok bool) {
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control#Cacheability
	return !strings.Contains(strings.ToLower(req.Header.Get(HeaderCacheControl)), "no-cache") ||
		!strings.Contains(strings.ToLower(req.Header.Get(HeaderCacheControl)), "no-store")
}

func requestMethodValid(req *http.Request) bool {
	return strings.ToLower(req.Method) == "get"
}
