package httpcache_test

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bxcodec/httpcache"
)

func Example_inMemoryStorageDefault() {
	client := &http.Client{}
	handler, err := httpcache.NewWithInmemoryCache(client, time.Second*15)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 100; i++ {
		startTime := time.Now()
		req, err := http.NewRequest("GET", "https://bxcodec.io", nil)
		if err != nil {
			log.Fatal((err))
		}
		res, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Response time: %v micro-second\n", time.Since(startTime).Microseconds())
		fmt.Println("Status Code", res.StatusCode)
		time.Sleep(time.Second * 1)
		fmt.Println("Sequence >>> ", i)
		if i%5 == 0 {
			err := handler.CacheInteractor.Flush()
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	// Example Output:
	/*
		2020/06/21 13:14:51 Cache item's missing failed to retrieve from cache, trying with a live version
		Response time: 940086 micro-second
		Status Code 200
		Sequence >>>  0
		2020/06/21 13:14:53 Cache item's missing failed to retrieve from cache, trying with a live version
		Response time: 73679 micro-second
		Status Code 200
		Sequence >>>  1
		Response time: 126 micro-second
		Status Code 200
		Sequence >>>  2
		Response time: 96 micro-second
		Status Code 200
		Sequence >>>  3
		Response time: 102 micro-second
		Status Code 200
		Sequence >>>  4
		Response time: 94 micro-second
		Status Code 200
		Sequence >>>  5
	*/
}
