package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/alacine/chu"
)

type User struct {
	Id    int    `validate:"id"`
	Name  string `validate:"word"`
	Email string `validate:"email"`
}

func hello(w http.ResponseWriter, r *http.Request) {
	name := chu.URLParam(r, "name")
	fmt.Fprintf(w, "hello, %s\n", name)
	user := &User{}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(user); err != nil {
		fmt.Fprintf(w, "%s\n", err)
		return
	}
	fmt.Println(user)
	if pass, err := chu.Validate(user); !pass {
		fmt.Fprintf(w, "%s\n", err)
		return
	}
	fmt.Fprintf(w, "User.id: %d\nUser.name: %s\nUser.email: %s\n", user.Id, user.Name, user.Email)
}

func main() {
	mux := chu.New()
	mux.HandleFunc(http.MethodGet, "/hello/:name", hello)
	log.Fatalln(http.ListenAndServe(":8200", mux))
}

//curl -X GET localhost:8200/hello/chu -d '{"id":1,"name":"--","email":"ryan@gmail.com"}'
//curl -X GET localhost:8200/hello/chu -d '{"id":1,"name":"chu","email":"ryan@gmail.com"}'
