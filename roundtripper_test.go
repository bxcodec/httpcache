package httpcache_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bxcodec/httpcache"
	"github.com/bxcodec/httpcache/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetFromCacheRoundtrip(t *testing.T) {
	mockCacheInteractor := new(mocks.ICacheInteractor)
	mockCacheInteractor.On("Get", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Once()
	client := &http.Client{}
	client.Transport = httpcache.NewRoundtrip(client.Transport, mockCacheInteractor)
	// HTTP GET 200
	jsonResp := []byte(`{"message": "Hello World!"}`)
	handler := func() (res http.Handler) {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/hello", r.RequestURI)
			w.Header().Set("Content-Type", "application/json")
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

	fmt.Println(">>> ", resp.Header.Get(httpcache.XHacheOrigin))

}
