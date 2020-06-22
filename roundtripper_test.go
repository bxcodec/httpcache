package httpcache_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bxcodec/httpcache"
	"github.com/bxcodec/httpcache/cache"
	"github.com/bxcodec/httpcache/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSetToCacheRoundtrip(t *testing.T) {
	mockCacheInteractor := new(mocks.ICacheInteractor)
	cachedResponse := cache.CachedResponse{}
	mockCacheInteractor.On("Get", mock.AnythingOfType("string")).Once().Return(cachedResponse, errors.New("uknown error"))
	mockCacheInteractor.On("Set", mock.AnythingOfType("string"), mock.Anything).Once().Return(nil)
	client := &http.Client{}
	client.Transport = httpcache.NewRoundtrip(http.DefaultTransport, mockCacheInteractor, true)
	// HTTP GET 200
	jsonResp := []byte(`{"message": "Hello World!"}`)
	handler := func() (res http.Handler) {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/hello", r.RequestURI)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "max-age=3600")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write(jsonResp)
			require.NoError(t, err)
		})
	}()

	mockServer := httptest.NewServer(handler)
	defer mockServer.Close()
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/hello", mockServer.URL), nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	require.Empty(t, resp.Header.Get(httpcache.XHacheOrigin))
	mockCacheInteractor.AssertExpectations(t)
}
