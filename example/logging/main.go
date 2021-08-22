package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/alacine/chu"
	"github.com/alacine/chu/middleware"
)

func main() {
	mux := chu.New()
	mux.Use(middleware.LogMiddleware)
	mux.HandleFunc(http.MethodGet, "/hello/:name", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(rw, "hello, %s\n", chu.URLParam(r, "name"))
	})
	log.Fatalln(http.ListenAndServe(":8000", mux))
}
