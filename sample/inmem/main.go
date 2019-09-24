package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bxcodec/hache"
)

func main() {
	client := &http.Client{}
	err := hache.NewWithInmemoryCache(client, time.Second*60)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		startTime := time.Now()
		req, err := http.NewRequest("GET", "https://bxcodec.io", nil)
		if err != nil {
			log.Fatal((err))
		}
		res, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Response time: %vms\n", time.Since(startTime).Microseconds())
		fmt.Println("Status Code", res.StatusCode)
	}
}
