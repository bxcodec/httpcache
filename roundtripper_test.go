package httpcache_test

import (
	"net/http"
	"testing"

	"github.com/bxcodec/httpcache"
)

func TestRoundtrip(t *testing.T) {
	client := &http.Client{}
	client.Transport = httpcache.NewRoundtrip(client.Transport, nil)

}
