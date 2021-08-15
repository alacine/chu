package chu

import (
	"fmt"
	"net/http"
	"strings"
)

var _ http.Handler = &Mux{}

// Mux 路由
type Mux struct {
	// 所有节点
	nodes []*node // MaxInt16

	// 邻接表，nex[i] = [j1, j2, ... jn] 表示节点 i 可以到达 j1, j2, ... jn
	next [][]int

	// map 中 key 为 http method，例如 GET, POST，value 为节点的编号
	//routers map[string]int
}

// New return a *Mux
func New() *Mux {
	cnt := 500
	nodes := make([]*node, 0, cnt)
	next := make([][]int, cnt)
	for i := range next {
		next[i] = make([]int, 0, cnt/50)
	}
	return &Mux{
		nodes: nodes,
		next:  next,
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
func (m *Mux) Handle(method, path string, handle func(http.ResponseWriter, *http.Request)) {
	if len(m.nodes) == 0 {
		m.nodes = append(m.nodes, &node{seg: "", level: 0})
	}
	//m.Show()
	addMethodToNode(method, path, handle, &m.nodes, &m.next)
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

// getHandleFunc 根据路径和 HTTP Method 匹配方法
// 如果找不到路径，handle 为 nil，c 为 1
// 如果找到路径，但对应的 HTTP Method 为 nil，则返回 handle 为 nil，c 为 2
func (m *Mux) getHandleFunc(method, path string) (http.HandlerFunc, int) {
	segs, err := pathToSegs(path)
	if err != nil {
		return nil, 3
	}
	idx := getLastMatchedNodeIdx(segs, m.nodes, m.next)
	mCode := methodMap[method]
	lastNode := m.nodes[idx]
	if lastNode.level != len(segs)-1 {
		return nil, 1
	}
	if lastNode.allowMethods&mCode == 0 {
		return nil, 2
	}
	return *lastNode.funcMap[mCode], 0
}

// ServeHTTP
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method, path := r.Method, r.URL.Path
	handle, code := m.getHandleFunc(method, path)
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
		handle(w, r)
	}
}
