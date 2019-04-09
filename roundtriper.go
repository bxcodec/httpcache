package hache

import "net/http"

// RoundTrip custom plugable' struct of implementation of the http.RoundTripper
type RoundTrip struct {
	DefaultRoundTripper http.RoundTripper
	CacheInteractor     CacheInteractor
}

// RoundTrip the implementation of http.RoundTripper
func (r *RoundTrip) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	panic("TODO: (bxcodec)")
	return
}

func validateCacheHeader() {
	panic("TODO: (bxcodec)")
}

func validateHTTPMethod() {
	panic("TODO: (bxcodec)")
}
