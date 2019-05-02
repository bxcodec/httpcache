package hache

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/bxcodec/hache/cache"
)

// Headers
const (
	HeaderAccept              = "Accept"
	HeaderAcceptEncoding      = "Accept-Encoding"
	HeaderAllow               = "Allow"
	HeaderAuthorization       = "Authorization"
	HeaderContentDisposition  = "Content-Disposition"
	HeaderContentEncoding     = "Content-Encoding"
	HeaderContentLength       = "Content-Length"
	HeaderContentType         = "Content-Type"
	HeaderCookie              = "Cookie"
	HeaderCacheControl        = "Cache-Control"
	HeaderSetCookie           = "Set-Cookie"
	HeaderIfModifiedSince     = "If-Modified-Since"
	HeaderLastModified        = "Last-Modified"
	HeaderLocation            = "Location"
	HeaderUpgrade             = "Upgrade"
	HeaderVary                = "Vary"
	HeaderWWWAuthenticate     = "WWW-Authenticate"
	HeaderXForwardedFor       = "X-Forwarded-For"
	HeaderXForwardedProto     = "X-Forwarded-Proto"
	HeaderXForwardedProtocol  = "X-Forwarded-Protocol"
	HeaderXForwardedSsl       = "X-Forwarded-Ssl"
	HeaderXUrlScheme          = "X-Url-Scheme"
	HeaderXHTTPMethodOverride = "X-HTTP-Method-Override"
	HeaderXRealIP             = "X-Real-IP"
	HeaderXRequestID          = "X-Request-ID"
	HeaderXRequestedWith      = "X-Requested-With"
	HeaderServer              = "Server"
	HeaderOrigin              = "Origin"

	// Access control
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"

	// Security
	HeaderStrictTransportSecurity = "Strict-Transport-Security"
	HeaderXContentTypeOptions     = "X-Content-Type-Options"
	HeaderXXSSProtection          = "X-XSS-Protection"
	HeaderXFrameOptions           = "X-Frame-Options"
	HeaderContentSecurityPolicy   = "Content-Security-Policy"
	HeaderXCSRFToken              = "X-CSRF-Token"
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

// RoundTrip the implementation of http.RoundTripper
func (r *RoundTrip) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if allowedFromCache(req) {
		resp, cachedItem, err := getCachedResponse(r.CacheInteractor, req)
		if resp != nil && err == nil {
			buildTheCachedResponseHeader(resp, cachedItem)
			return resp, err
		}
	}

	resp, err = r.DefaultRoundTripper.RoundTrip(req)
	if err != nil {
		return
	}

	if !allowedToCache(req, resp) {
		fmt.Println("MASUK PAKDE>>>>")
		return
	}
	fmt.Println("Stored to cache")
	storeRespToCache(r.CacheInteractor, req, resp)
	return
}

func storeRespToCache(cacheInteractor cache.Interactor, req *http.Request, resp *http.Response) (err error) {
	cachedResp := cache.CachedResponse{
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
	cachedResp.DumpedBody = bodyBytes
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
		fmt.Println("FAILED HERE 2", ok)
		return
	}

	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control#Preventing_caching
	if ok = strings.ToLower(req.Header.Get(HeaderCacheControl)) != "no-store"; !ok {
		fmt.Println("FAILED HERE 3")
		return
	}
	if ok = strings.ToLower(resp.Header.Get(HeaderCacheControl)) != "no-store"; !ok {
		fmt.Println("FAILED HERE 4")
		return
	}

	// Only cache the response of with code 200
	if ok = resp.StatusCode == http.StatusOK; !ok {
		fmt.Println("FAILED HERE 4")
		return
	}
	fmt.Println("FAILED HERE ", ok)
	return
}

func allowedFromCache(req *http.Request) (ok bool) {
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control#Cacheability
	return strings.ToLower(req.Header.Get(HeaderCacheControl)) != "no-cache"
}

func requestMethodValid(req *http.Request) bool {
	fmt.Println("Method >>", req.Method == http.MethodGet, strings.ToLower(req.Method) == "get")
	return req.Method == http.MethodGet || strings.ToLower(req.Method) == "get"
}
