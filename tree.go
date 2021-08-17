package chu

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type methodType uint

const (
	mGET methodType = 1 << iota
	mPOST
	mPUT
	mDELETE
	mHEAD
	mOPTION
	mPATCH
	mCONNECT
	mTRACE
)

var methodMap = map[string]methodType{
	http.MethodGet:     mGET,
	http.MethodPost:    mPOST,
	http.MethodPut:     mPUT,
	http.MethodDelete:  mDELETE,
	http.MethodHead:    mHEAD,
	http.MethodOptions: mOPTION,
	http.MethodPatch:   mPATCH,
	http.MethodConnect: mCONNECT,
	http.MethodTrace:   mTRACE,
}

type nodemethods map[uint]*http.HandlerFunc

type node struct {
	seg          string
	wildcard     bool
	wildchild    bool
	level        int
	allowMethods methodType
	funcMap      map[methodType]*ChuHandlerFunc
}

// dfs 按照深度优先顺序打出所有可用路由
func dfs(idx int, nodes []*node, nex [][]int, printSegs []string) {
	if ams := nodes[idx].allowMethods; ams != 0 {
		//fmt.Printf("%d %#v %v\n", idx, strings.Join(printSegs, "/"), nex[idx])
		mList := make([]string, 0, len(methodMap))
		for k, v := range methodMap {
			if ams&v != 0 {
				mList = append(mList, k)
			}
		}
		fmt.Printf("%d %#v %v\n", idx, strings.Join(printSegs, "/"), mList)
	}
	for _, i := range nex[idx] {
		printSegs = append(printSegs, nodes[i].seg)
		l := len(printSegs)
		dfs(i, nodes, nex, printSegs)
		printSegs = printSegs[:l-1]
	}
	printSegs = []string{""}
}

// addMethodNode 添加一个节点
// path: 完整的注册路径
// nodes: 所有节点
// nex: 节点邻接表
func addMethodToNode(method string, path string, handle ChuHandlerFunc, nodes *[]*node, nex *[][]int) {
	segs, err := pathToSegs(path)
	if err != nil {
		panic(err)
	}
	idx := getLastMatchedNodeIdx(segs, *nodes, *nex)
	mCode, ok := methodMap[method]
	if !ok {
		panic("No such HTTP Method called: " + method)
	}
	lastNode := (*nodes)[idx]
	if lastNode.level == len(segs)-1 && lastNode.allowMethods&mCode != 0 {
		panic("Already have handle func for " + path + " with " + method)
	}
	if (*lastNode).level < len(segs)-1 {
		// si 是 segs 中第一个不匹配的序号
		si := lastNode.level + 1

		// 通配符类型节点作为子节点时，该节点只允许有一个子节点
		if len((*nex)[idx]) != 0 && (lastNode.wildchild || isWildcard(segs[si])) {
			panic("Conflict between " + path + " and " +
				strings.Join(segs[:si], "/") + "/" +
				(*nodes)[(*nex)[idx][0]].seg)
		}
		idx = createNodes(idx, segs[si:], nodes, nex)
		lastNode = (*nodes)[idx]
	}
	lastNode.allowMethods |= mCode
	if lastNode.funcMap == nil {
		lastNode.funcMap = make(map[methodType]*ChuHandlerFunc)
	}
	lastNode.funcMap[mCode] = &handle
}

// getLastMatchedNodeIdx 返回最后一个匹配的节点
func getLastMatchedNodeIdx(segs []string, nodes []*node, nex [][]int) int {
	if len(segs) == 0 {
		return 0
	}
	// a, b 只是 si 和 idx 的一个备份，用来检测 si 和 idx 是否发生变化
	si, idx, a, b := 1, 0, 1, 0
	for si < len(segs) {
		a, b = si, idx
		for _, i := range nex[idx] {
			if nodes[i].seg == segs[si] {
				si, idx = si+1, i
				break
			}
		}
		if a == si || b == idx {
			break
		}
	}
	return idx
}

// createNodes 在某一节点后创建新的节点链，并且返回最后一个节点的编号
// from: 第一个不匹配的节点编号
// segs: 需要新添加的段
func createNodes(from int, segs []string, nodes *[]*node, nex *[][]int) int {
	pre := (*nodes)[from]
	for _, seg := range segs {
		isWild := isWildcard(seg)
		t := &node{
			seg:      seg,
			wildcard: isWild,
			level:    pre.level + 1,
		}
		ti := len(*nodes)
		*nodes = append(*nodes, t)
		(*nex)[from] = append((*nex)[from], ti)
		pre.wildchild = isWild
		pre = t
		from = ti
	}
	return from
}

// isWildcard 判断是否为通配符类型的节点
func isWildcard(seg string) bool {
	if len(seg) > 0 && seg[0] == ':' {
		return true
	}
	return false
}

// pathToSegs 把路径以斜线 '/' 为分割符号拆成多段
// 不允许出现 "//"、":/"、"::"、"/:xxxxx:xxxx/" 这种类型，但未尾可以有 "//"、"///" 等
func pathToSegs(path string) ([]string, error) {
	path, err := trimSlash(path)
	if err != nil {
		return nil, err
	}
	n := len(path)
	if n > 0 && path[n-1] == ':' {
		return nil, errors.New("Invalid path")
	}
	last := byte('/')
	for i := 1; i < n; i++ {
		a, b := path[i-1], path[i]
		if b != ':' && b != '/' {
			continue
		}
		if b == ':' {
			if last == b {
				return nil, errors.New("Invalid path")
			}
		} else {
			if a == '/' || a == ':' {
				return nil, errors.New("Invalid path")
			}
		}
		last = b
	}
	return strings.Split(path, "/"), nil
}

// trimSlash 去掉末尾 '/'
func trimSlash(path string) (string, error) {
	if len(path) == 0 {
		return "", errors.New("path should not be empty")
	}
	if path[0] != '/' {
		return "", errors.New("path should start with '/'")
	}
	return strings.TrimRight(path, "/"), nil
}
