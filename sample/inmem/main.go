package main

import (
	"fmt"
	"io/ioutil"
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
		req, err := http.NewRequest("GET", "https://google.com", nil)
		if err != nil {
			log.Fatal((err))
		}
		res, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Response time: %v micro-second\n", time.Since(startTime).Microseconds())
		fmt.Println("Status Code", res.StatusCode)
		fmt.Println("Header", res.Header)
		// fmt.Println("Header", res.Header.Get("expires"))
		// printBody(res)
	}
}

func printBody(resp *http.Response) {
	jbyt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ResponseBody: \t%s\n", string(jbyt))
}
