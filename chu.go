package chu

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

var _ http.Handler = &Mux{}

// URL path 中的参数
type Params map[string]string

func (ps *Params) ByName(name string) string {
	return (*ps)[name]
}

type ChuHandlerFunc func(http.ResponseWriter, *http.Request, *Params)

// Mux 路由
type Mux struct {
	// 所有节点
	nodes []*node // MaxInt16

	// 邻接表，next[i] = [j1, j2, ... jn] 表示节点 i 可以到达 j1, j2, ... jn
	next [][]int

	// URL 参数池，当从 URL 中获取到参数时，从这里面拿内存来存放参数
	// 避免多次分配内存，这个做法是从 httprouter 里学来的
	paramPool sync.Pool
}

// New return a *Mux
func New() *Mux {
	cnt := 500
	nodes := make([]*node, 0, cnt)
	next := make([][]int, cnt)
	for i := range next {
		next[i] = make([]int, 0, cnt/50)
	}
	paramPool := sync.Pool{
		New: func() interface{} {
			return &Params{}
		},
	}
	return &Mux{
		nodes:     nodes,
		next:      next,
		paramPool: paramPool,
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
func (m *Mux) Handle(method, path string, handle func(http.ResponseWriter, *http.Request, *Params)) {
	if len(m.nodes) == 0 {
		m.nodes = append(m.nodes, &node{seg: "", level: 0})
	}
	//m.Show()
	addMethodToNode(method, path, ChuHandlerFunc(handle), &m.nodes, &m.next)
}

// Get Handle
func (m *Mux) Get(path string, handle func(http.ResponseWriter, *http.Request, *Params)) {
	m.Handle(http.MethodGet, path, handle)
}

// Post Handle
func (m *Mux) Post(path string, handle func(http.ResponseWriter, *http.Request, *Params)) {
	m.Handle(http.MethodPost, path, handle)
}

// Delete Handle
func (m *Mux) Delete(path string, handle func(http.ResponseWriter, *http.Request, *Params)) {
	m.Handle(http.MethodDelete, path, handle)
}

// Put Handle
func (m *Mux) Put(path string, handle func(http.ResponseWriter, *http.Request, *Params)) {
	m.Handle(http.MethodPut, path, handle)
}

// Head Handle
func (m *Mux) Head(path string, handle func(http.ResponseWriter, *http.Request, *Params)) {
	m.Handle(http.MethodHead, path, handle)
}

// findMatchedNode 返回根据 http method 和 URL path 匹配到的节点的编号，
// 同时获取 URL path 中的参数
// 与 getLastMatchedNodeIdx 不同，findMatchedNode 支持具体的参数和通配类型节点匹配
func (m *Mux) findMatchedNodeWithParam(method, path string) (idx int, ps *Params) {
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
				if ps == nil {
					ps, _ = m.paramPool.Get().(*Params)
				}
				(*ps)[curNode.seg[1:]] = segs[si]
				si, idx = si+1, i
				break
			}
			if m.nodes[i].seg == segs[si] {
				si, idx = si+1, i
				break
			}
		}
		if a == si || b == idx {
			if ps != nil {
				m.paramPool.Put(ps)
			}
			return -1, nil
		}
	}
	return idx, ps
}

// getHandleFunc 根据路径和 HTTP Method 匹配方法
// 如果找不到路径，handle 为 nil，c 为 1
// 如果找到路径，但对应的 HTTP Method 为 nil，则返回 handle 为 nil，c 为 2
func (m *Mux) getHandleFunc(method, path string) (ChuHandlerFunc, *Params, int) {
	idx, ps := m.findMatchedNodeWithParam(method, path)
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
	handle, ps, code := m.getHandleFunc(method, path)
	if ps != nil {
		defer m.paramPool.Put(ps)
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
		handle(w, r, ps)
	}
}
