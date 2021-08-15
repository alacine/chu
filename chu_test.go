package chu

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func fakeHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {}
}

func TestHandle(t *testing.T) {
	tests := []struct {
		path, method string
		handlefunc   func(http.ResponseWriter, *http.Request)
	}{
		{
			path:       "/api",
			method:     "GET",
			handlefunc: fakeHandler(),
		},
		{
			path:       "/api/:name",
			method:     "GET",
			handlefunc: fakeHandler(),
		},
		{
			path:       "/api",
			method:     "POST",
			handlefunc: fakeHandler(),
		},
		{
			path:       "/book",
			method:     "GET",
			handlefunc: fakeHandler(),
		},
		{
			path:       "/book/:id/info",
			method:     "GET",
			handlefunc: fakeHandler(),
		},
		{
			path:       "/book/:id/info",
			method:     "POST",
			handlefunc: fakeHandler(),
		},
		{
			path:       "/book/:id",
			method:     "DELETE",
			handlefunc: fakeHandler(),
		},
	}
	mux := New()
	for _, test := range tests {
		mux.Handle(test.method, test.path, test.handlefunc)
	}
	mux.Show()
}

func TestMux(t *testing.T) {
	mux := New()
	mux.Handle("GET", "/ping", func(rw http.ResponseWriter, r *http.Request) {
		io.WriteString(rw, "pong")
	})

	rw := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/book", nil)
	mux.ServeHTTP(rw, req)
	if code := rw.Result().StatusCode; code != http.StatusNotFound {
		t.Errorf("expect 404 status code, got %v", code)
	}

	rw1 := httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/ping", nil)
	mux.ServeHTTP(rw1, req)
	if code := rw1.Result().StatusCode; code != http.StatusOK {
		t.Errorf("expect 200 status code, got %v", code)
	}

	rw2 := httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/ping", nil)
	mux.ServeHTTP(rw2, req)
	if code := rw2.Result().StatusCode; code != http.StatusMethodNotAllowed {
		t.Errorf("expect 405 status code, got %v", code)
	}
}
