package chu

import (
	"fmt"
	"net/http"
)

var _ http.Handler = &Mux{}

// Mux 路由
type Mux struct {
	routers map[string]map[string]http.HandlerFunc
}

// New return a *Mux
func New() *Mux {
	r := make(map[string]map[string]http.HandlerFunc)
	return &Mux{routers: r}
}

// Show 打印所有路由
func (m *Mux) Show() {
	for k, v := range m.routers {
		fmt.Printf("Path: %v \t HandleFunc: %v\n", k, v)
	}
}

// Handle 注册路由
func (m *Mux) Handle(method, path string, handle func(http.ResponseWriter, *http.Request)) {
	_, ok1 := m.routers[path]
	_, ok2 := m.routers[path][method]
	if ok1 && ok2 {
		panic("registe router conflict")
	}
	if !ok1 {
		m.routers[path] = make(map[string]http.HandlerFunc)
	}
	m.routers[path][method] = handle
}

// Get Handle
func (m *Mux) Get(path string, handle func(http.ResponseWriter, *http.Request)) {
	m.Handle(http.MethodGet, path, handle)
}

// Post Handle
func (m *Mux) Post(path string, handle func(http.ResponseWriter, *http.Request)) {
	m.Handle(http.MethodPost, path, handle)
}

// Delete Handle
func (m *Mux) Delete(path string, handle func(http.ResponseWriter, *http.Request)) {
	m.Handle(http.MethodDelete, path, handle)
}

// Put Handle
func (m *Mux) Put(path string, handle func(http.ResponseWriter, *http.Request)) {
	m.Handle(http.MethodPut, path, handle)
}

// Head Handle
func (m *Mux) Head(path string, handle func(http.ResponseWriter, *http.Request)) {
	m.Handle(http.MethodHead, path, handle)
}

func (m *Mux) getHandleFunc(method, path string) (http.HandlerFunc, int) {
	_, ok1 := m.routers[path]
	h, ok2 := m.routers[path][method]
	if ok1 && ok2 {
		return h, 0
	}
	if !ok1 {
		return nil, 1
	}
	return nil, 2
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Show()
	method, path := r.Method, r.URL.Path
	handle, errCode := m.getHandleFunc(method, path)
	switch errCode {
	case 1:
		http.NotFound(w, r)
	case 2:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	default:
		handle(w, r)
	}
}
