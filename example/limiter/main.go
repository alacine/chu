package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alacine/chu"
	"github.com/alacine/chu/middleware"
)

func hello(w http.ResponseWriter, r *http.Request) {
	name := chu.URLParam(r, "name")
	time.Sleep(time.Second * 1)
	fmt.Fprintf(w, "hello, %s\n", name)
}

func main() {
	mux := chu.New()
	//mux.Use(middleware.Limiter(10))
	// 最大突发并发数为 100，每 500 毫秒补充一个 token
	mux.Use(middleware.BurstBucketLimiter(100, time.Millisecond*20))
	mux.Use(middleware.RequestID)
	mux.Use(middleware.LogMiddleware)
	mux.Get("/hello/:name", hello)
	log.Fatalln(http.ListenAndServe(":8003", mux))
}
