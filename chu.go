// Package chu ...
//
// Example:
//  package main
//
//  import (
//  	"fmt"
//  	"log"
//  	"net/http"
//
//  	"github.com/alacine/chu"
//  )
//
//  func hello(w http.ResponseWriter, r *http.Request) {
//  	name := chu.URLParam(r, "name")
//  	fmt.Fprintf(w, "hello, %s\n", name)
//  }
//
//  func main() {
//  	mux := chu.New()
//  	mux.Get("/hello/:name", hello)
//  	log.Fatalln(http.ListenAndServe(":8000", mux))
//  }
package chu

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

var _ http.Handler = &Mux{}

// Middleware 中间件
type Middleware func(http.Handler) http.Handler

// Mux 路由
type Mux struct {
	// 所有节点
	nodes []*node // MaxInt16

	// 邻接表，next[i] = [j1, j2, ... jn] 表示节点 i 可以到达 j1, j2, ... jn
	next [][]int

	// Context 池，当从 URL 中获取到参数时，从这里面拿 Context 来存放参数
	// 避免多次分配内存
	contextPool *sync.Pool

	// 中间件
	middlewares []Middleware
}

// New return a *Mux
func New() *Mux {
	cnt := 500
	nodes := make([]*node, 0, cnt)
	next := make([][]int, cnt)
	for i := range next {
		next[i] = make([]int, 0, cnt/50)
	}
	contextPool := sync.Pool{
		New: func() interface{} {
			return NewChuContext()
		},
	}
	return &Mux{
		nodes:       nodes,
		next:        next,
		contextPool: &contextPool,
	}
}

// Show 打印所有可用路由
func (m *Mux) Show() {
	fmt.Printf("len(m.nodes): %v\n", len(m.nodes))
	for i, n := range m.nodes {
		fmt.Printf("idx(%v): ", i)
		indent := strings.Repeat("  ", n.level)
		fmt.Printf("%v%#v allowMethods: %d, ", indent, n.seg, n.allowMethods)
		fmt.Printf("n.wildcard: %v, ", n.wildcard)
		fmt.Printf("n.wildchild: %v\n", n.wildchild)
	}

	if len(m.nodes) > 0 {
		dfs(0, m.nodes, m.next, []string{""})
	}
}

// Handle 注册路由
func (m *Mux) handle(method, path string, handler http.Handler) {
	if len(m.nodes) == 0 {
		m.nodes = append(m.nodes, &node{seg: "", level: 0})
	}
	for i := len(m.middlewares) - 1; i >= 0; i-- {
		handler = m.middlewares[i](handler)
	}
	addMethodToNode(method, path, handler, &m.nodes, &m.next)
}

func (m *Mux) Handle(method, path string, handler http.Handler) {
	m.handle(method, path, handler)
}

func (m *Mux) HandleFunc(method, path string, handle http.HandlerFunc) {
	m.handle(method, path, handle)
}

// Use 为 Mux 添加中间件
func (m *Mux) Use(middlewares ...Middleware) {
	if m.middlewares == nil {
		m.middlewares = make([]Middleware, 0, len(middlewares))
	}
	m.middlewares = append(m.middlewares, middlewares...)
}

// Get HandleFunc
func (m *Mux) Get(path string, handle http.HandlerFunc) {
	m.HandleFunc(http.MethodGet, path, handle)
}

// Post HandleFunc
func (m *Mux) Post(path string, handle http.HandlerFunc) {
	m.HandleFunc(http.MethodPost, path, handle)
}

// Delete HandleFunc
func (m *Mux) Delete(path string, handle http.HandlerFunc) {
	m.HandleFunc(http.MethodDelete, path, handle)
}

// Put HandleFunc
func (m *Mux) Put(path string, handle http.HandlerFunc) {
	m.HandleFunc(http.MethodPut, path, handle)
}

// Head HandleFunc
func (m *Mux) Head(path string, handle http.HandlerFunc) {
	m.HandleFunc(http.MethodHead, path, handle)
}

// findMatchedNode 返回根据 http method 和 URL path 匹配到的节点的编号和 Context
// Context 中有从 URL path 中获取的参数
// 与 getLastMatchedNodeIdx 不同，findMatchedNode 支持具体的参数和通配类型节点匹配
func (m *Mux) findMatchedNode(method, path string) (idx int, ctx *Context) {
	path = strings.TrimRight(path, "/")
	segs := strings.Split(path, "/")
	n := len(segs)
	if n == 0 {
		return 0, nil
	}

	// a, b 只是 si 和 idx 的一个备份，用来检测 si 和 idx 是否发生变化
	si, idx, a, b := 1, 0, 1, 0
	for si < len(segs) {
		a, b = si, idx
		for _, i := range m.next[idx] {
			curNode := m.nodes[i]
			if curNode.wildcard {
				if ctx == nil {
					ctx, _ = m.contextPool.Get().(*Context)
				}
				ctx.URLParams.Keys = append(ctx.URLParams.Keys, curNode.seg[1:])
				ctx.URLParams.Values = append(ctx.URLParams.Values, segs[i])
				si, idx = si+1, i
				break
			}
			if m.nodes[i].seg == segs[si] {
				si, idx = si+1, i
				break
			}
		}
		if a == si || b == idx {
			if ctx != nil {
				m.contextPool.Put(ctx)
			}
			return -1, nil
		}
	}
	return idx, ctx
}

// getHandler 根据路径和 HTTP Method 匹配方法，同时返回 Context 和匹配状态码
// 如果找不到路径，返回的 handler 为 nil，状态码为 1
// 如果找到路径，但对应的 HTTP Method 为 nil，则返回 handle 为 nil，状态码为 2
func (m *Mux) getHandler(method, path string) (http.Handler, *Context, int) {
	idx, ps := m.findMatchedNode(method, path)
	if idx == -1 {
		return nil, ps, 1
	}
	mCode := methodMap[method]
	lastNode := m.nodes[idx]
	if lastNode.allowMethods&mCode == 0 {
		return nil, ps, 2
	}
	return *lastNode.funcMap[mCode], ps, 0
}

// ServeHTTP
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method, path := r.Method, r.URL.Path
	handle, ctx, code := m.getHandler(method, path)
	if ctx != nil {
		r = r.WithContext(context.WithValue(r.Context(), ContextKey, ctx))
		defer m.contextPool.Put(ctx)
	}
	switch code {
	case 1:
		http.NotFound(w, r)
	case 2:
		http.Error(
			w,
			http.StatusText(http.StatusMethodNotAllowed),
			http.StatusMethodNotAllowed,
		)
	case 3:
		http.Error(
			w,
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest,
		)
	default:
		handle.ServeHTTP(w, r)
	}
}
