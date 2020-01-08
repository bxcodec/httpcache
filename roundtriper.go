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
	cacheControl "github.com/bxcodec/hache/control/cacheheader"
)

// Headers
const (
	HeaderAuthorization = "Authorization"
	HeaderCacheControl  = "Cache-Control"
	// To indicate that the response is got from this hache library
	XFromHache   = "X-Hache"
	XHacheOrigin = "X-Hache-Origin"
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

func validateTheCacheControl(req *http.Request, resp *http.Response) (validationResult cacheControl.ObjectResults, err error) {
	reqDir, err := cacheControl.ParseRequestCacheControl(req.Header.Get("Cache-Control"))
	if err != nil {
		return
	}

	resDir, err := cacheControl.ParseResponseCacheControl(resp.Header.Get("Cache-Control"))
	if err != nil {
		return
	}

	expiry := resp.Header.Get("Expires")
	expiresHeader, err := http.ParseTime(expiry)
	if err != nil && expiry != "" &&
		// https://stackoverflow.com/questions/11357430/http-expires-header-values-0-and-1
		expiry != "-1" && expiry != "0" {
		return
	}

	dateHeaderStr := resp.Header.Get("Date")
	dateHeader, err := http.ParseTime(dateHeaderStr)
	if err != nil && dateHeaderStr != "" {
		return
	}

	lastModifiedStr := resp.Header.Get("Last-Modified")
	lastModifiedHeader, err := http.ParseTime(lastModifiedStr)
	if err != nil && lastModifiedStr != "" {
		return
	}

	obj := cacheControl.Object{
		RespDirectives:         resDir,
		RespHeaders:            resp.Header,
		RespStatusCode:         resp.StatusCode,
		RespExpiresHeader:      expiresHeader,
		RespDateHeader:         dateHeader,
		RespLastModifiedHeader: lastModifiedHeader,
		ReqDirectives:          reqDir,
		ReqHeaders:             req.Header,
		ReqMethod:              req.Method,
		NowUTC:                 time.Now().UTC(),
	}

	validationResult = cacheControl.ObjectResults{}
	cacheControl.CachableObject(&obj, &validationResult)
	cacheControl.ExpirationObject(&obj, &validationResult)

	return validationResult, nil
}

// RoundTrip the implementation of http.RoundTripper
func (r *RoundTrip) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if allowedFromCache(req.Header) {
		resp, cachedItem, err := getCachedResponse(r.CacheInteractor, req)
		if resp != nil && err == nil {
			buildTheCachedResponseHeader(resp, cachedItem, r.CacheInteractor.Origin())
			return resp, err
		}
	}

	err = nil
	resp, err = r.DefaultRoundTripper.RoundTrip(req)
	if err != nil {
		return
	}

	// Only cache the response of with Success Status
	if resp.StatusCode >= http.StatusMultipleChoices ||
		resp.StatusCode < http.StatusOK ||
		resp.StatusCode == http.StatusNoContent {
		return
	}

	validationResult, err := validateTheCacheControl(req, resp)
	if err != nil {
		return
	}

	if validationResult.OutErr != nil {
		return
	}

	// reasons to not to cache
	if len(validationResult.OutReasons) > 0 {
		return
	}

	err = storeRespToCache(r.CacheInteractor, req, resp)
	if err != nil {
		log.Printf("Can't store the response to database, plase check. Err: %v\n", err)
		err = nil // set err back to nil to make the call still success.
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

	err = cacheInteractor.Set(getCacheKey(req), cachedResp, 0)
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

	validationResult, err := validateTheCacheControl(req, resp)
	if err != nil {
		return
	}

	if validationResult.OutErr != nil {
		return
	}

	if time.Now().After(validationResult.OutExpirationTime) {
		err = fmt.Errorf("cached-item already expired")
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
func buildTheCachedResponseHeader(resp *http.Response, cachedResp cache.CachedResponse, origin string) {
	resp.Header.Add("Expires", cachedResp.CachedTime.String())
	resp.Header.Add(XFromHache, "true")
	resp.Header.Add(XHacheOrigin, origin)
	// TODO: (bxcodec) add more headers related to cache
}

func allowedToCache(header http.Header, method string) (ok bool) {
	// A request with authorization header must not be cached
	// https://tools.ietf.org/html/rfc7234#section-3.2
	// Unless configured by user to cache request by authorization
	if ok = (!CacheAuthorizedRequest && header.Get(HeaderAuthorization) == ""); !ok {
		return
	}

	// check if the request method allowed to be cached
	if strings.ToLower(method) != "get" {
		return
	}

	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control#Preventing_caching
	if ok = strings.ToLower(header.Get(HeaderCacheControl)) != "no-store"; !ok {
		return
	}
	if ok = strings.ToLower(header.Get(HeaderCacheControl)) != "no-store"; !ok {
		return
	}

	return true
}

func allowedFromCache(header http.Header) (ok bool) {
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control#Cacheability
	return !strings.Contains(strings.ToLower(header.Get(HeaderCacheControl)), "no-cache") ||
		!strings.Contains(strings.ToLower(header.Get(HeaderCacheControl)), "no-store")
}
