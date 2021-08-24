package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alacine/chu"
	"github.com/alacine/chu/middleware"
)

func helloWithTimeout(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	select {
	case <-ctx.Done():
		return
	case <-time.After(time.Second * 2):
		fmt.Fprintf(rw, "hello, %s\n", chu.URLParam(r, "name"))
	}
	log.Println("done")
}

func main() {
	mux := chu.New()
	mux.Use(middleware.Timeout(time.Second * 1))
	mux.Use(middleware.RequestID)
	mux.Use(middleware.LogMiddleware)
	mux.HandleFunc(http.MethodGet, "/hello/:name", helloWithTimeout)
	log.Fatalln(http.ListenAndServe(":8001", mux))
}
