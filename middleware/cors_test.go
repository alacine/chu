package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alacine/chu"
)

func TestCorsMiddleware(t *testing.T) {
	mux := chu.New()
	mux.Use(CorsMiddleware)
	mux.HandleFunc(http.MethodGet, "/hello/:name", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(rw, "hello,", chu.URLParam(r, "name"))
	})

	rw := httptest.NewRecorder()
	p1 := "chu"
	req, _ := http.NewRequest(http.MethodGet, "/hello/"+p1, nil)
	mux.ServeHTTP(rw, req)
	defer rw.Result().Body.Close()
	acao := rw.Result().Header.Get("Access-Control-Allow-Origin")
	acam := rw.Result().Header.Get("Access-Control-Allow-Methods")
	t.Log(acao, acam)
	acao_want, acam_want := "*", "GET, POST, DELETE"
	if acao != acao_want {
		t.Errorf("expect %v, but get %v", acao_want, acao)
	}
	if acam != acam_want {
		t.Errorf("expect %v, but get %v", acam_want, acam)
	}
}
