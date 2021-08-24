package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alacine/chu"
	"github.com/alacine/chu/middleware"
)

func main() {
	mux := chu.New()
	mux.Use(middleware.Timeout(time.Second * 1))
	mux.Use(middleware.RequestID)
	mux.Use(middleware.LogMiddleware)
	mux.HandleFunc(http.MethodGet, "/hello/:name", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(rw, "hello, %s\n", chu.URLParam(r, "name"))
	})
	log.Fatalln(http.ListenAndServe(":8000", mux))
}
