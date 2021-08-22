package chu

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func fakeHandlerFunc() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {}
}

func catchPanic(f func()) (rec interface{}) {
	defer func() {
		rec = recover()
	}()
	f()
	return
}

type testRouter struct {
	path, method string
	handlefunc   http.HandlerFunc
}

func TestHandle(t *testing.T) {
	// 正常的路由注册
	tests := []testRouter{
		{
			path:       "/api",
			method:     "GET",
			handlefunc: fakeHandlerFunc(),
		},
		{
			path:       "/",
			method:     "GET",
			handlefunc: fakeHandlerFunc(),
		},
		{
			path:       "/api/:name",
			method:     "GET",
			handlefunc: fakeHandlerFunc(),
		},
		{
			path:       "/api",
			method:     "POST",
			handlefunc: fakeHandlerFunc(),
		},
		{
			path:       "/book",
			method:     "GET",
			handlefunc: fakeHandlerFunc(),
		},
		{
			path:       "/book/:id/info",
			method:     "GET",
			handlefunc: fakeHandlerFunc(),
		},
		{
			path:       "/book/:id/info",
			method:     "POST",
			handlefunc: fakeHandlerFunc(),
		},
		{
			path:       "/book/:id",
			method:     "DELETE",
			handlefunc: fakeHandlerFunc(),
		},
	}
	mux := New()
	for _, test := range tests {
		t.Logf("test.method: %v, ", test.method)
		t.Logf("test.path: %v\n", test.path)
		mux.Handle(test.method, test.path, test.handlefunc)
	}

	// 会发生冲突 panic 的路由注册
	conflictTests := []testRouter{
		{
			path:       "/api/:id",
			method:     "GET",
			handlefunc: fakeHandlerFunc(),
		},
		{
			path:       "/api/abc",
			method:     "POST",
			handlefunc: fakeHandlerFunc(),
		},
		{
			path:       "/book/info",
			method:     "GET",
			handlefunc: fakeHandlerFunc(),
		},
		{
			path:       "/book/:ids",
			method:     "GET",
			handlefunc: fakeHandlerFunc(),
		},
	}
	conflictPanicReg := regexp.MustCompile(`Conflict between .* and .*`)
	for _, ct := range conflictTests {
		rec := catchPanic(func() {
			mux.Handle(ct.method, ct.path, ct.handlefunc)
		})
		if rec == nil {
			t.Errorf("should panic but not: %s %s", ct.method, ct.path)
		}
		recstr := fmt.Sprint(rec)
		matched := conflictPanicReg.MatchString(recstr)
		if !matched {
			t.Errorf("got panic msg: '%s', but want 'Conflict between ... and ...'", recstr)
		}
	}

	// 会发生重复注册 panic 的路由注册
	repeatedTests := []testRouter{
		{
			path:       "/api/:name",
			method:     "GET",
			handlefunc: fakeHandlerFunc(),
		},
	}
	repeatedPanicReg := regexp.MustCompile(`Already have handle func for .* with .*`)
	for _, rt := range repeatedTests {
		rec := catchPanic(func() {
			mux.Handle(rt.method, rt.path, rt.handlefunc)
		})
		if rec == nil {
			t.Errorf("should panic but not: %s %s", rt.method, rt.path)
		}
		recstr := fmt.Sprint(rec)
		matched := repeatedPanicReg.MatchString(recstr)
		if !matched {
			t.Errorf("got panic msg: '%s', but want 'Already have handle func for ... with ...'", recstr)
		}
	}

	// 会发生找不到对应 HTTP Method panic 的路由注册
	noHTTPMethodTests := []testRouter{
		{
			path:       "/api/:name",
			method:     "GETT",
			handlefunc: fakeHandlerFunc(),
		},
	}
	noMethodPanicReg := regexp.MustCompile(`No such HTTP Method called: .*`)
	for _, nt := range noHTTPMethodTests {
		rec := catchPanic(func() {
			mux.Handle(nt.method, nt.path, nt.handlefunc)
		})
		if rec == nil {
			t.Errorf("should panic but not: %s %s", nt.method, nt.path)
		}
		recstr := fmt.Sprint(rec)
		matched := noMethodPanicReg.MatchString(recstr)
		if !matched {
			t.Errorf("got panic msg: '%s', but want 'No such HTTP Method called: ...'", recstr)
		}
	}
	mux.Show()
}

func TestMux(t *testing.T) {
	p1, p2 := "id", "name"
	mux := New()
	mux.HandleFunc(
		"GET",
		"/ping/:"+p1+"/info/:"+p2,
		func(rw http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(rw, "pong")
			fmt.Fprintln(rw, p1, URLParam(r, p1))
			fmt.Fprintln(rw, p2, URLParam(r, p2))
		},
	)

	rw := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/book", nil)
	mux.ServeHTTP(rw, req)
	if code := rw.Result().StatusCode; code != http.StatusNotFound {
		t.Errorf("expect 404 status code, got %v", code)
	}
	defer rw.Result().Body.Close()

	rw = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/ping/abc/info/hah", nil)
	mux.ServeHTTP(rw, req)
	if code := rw.Result().StatusCode; code != http.StatusOK {
		t.Errorf("expect 200 status code, got %v", code)
	}
	resp, _ := ioutil.ReadAll(rw.Result().Body)
	respStr := string(resp)
	if expResp := fmt.Sprintf("%s\n%s %s\n%s %s\n", "pong", p1, "abc", p2, "hah"); respStr != expResp {
		t.Errorf("expect %#v , got %#v", expResp, respStr)
	}

	rw = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/ping", nil)
	mux.ServeHTTP(rw, req)
	if code := rw.Result().StatusCode; code != http.StatusMethodNotAllowed {
		t.Errorf("expect 405 status code, got %v", code)
	}
}
