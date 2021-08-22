package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/alacine/chu"
)

func hello(w http.ResponseWriter, r *http.Request) {
	name := chu.URLParam(r, "name")
	fmt.Fprintf(w, "hello, %s\n", name)
}

func main() {
	mux := chu.New()
	mux.Get("/hello/:name", hello)
	log.Fatalln(http.ListenAndServe(":8000", mux))
}
