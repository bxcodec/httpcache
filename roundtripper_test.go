package hache_test

import (
	"net/http"
	"testing"

	"github.com/bxcodec/hache"
)

func TestRoundtrip(t *testing.T) {
	client := &http.Client{}
	client.Transport = hache.NewRoundtrip(client.Transport, nil)
}
