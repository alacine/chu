package chu

import (
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
	funcMap      map[methodType]*http.HandlerFunc
}

// dfs 按照深度优先顺序打出所有可用路由
func dfs(idx int, allNodes []*node, nex [][]int, printSegs []string) {
	if ams := allNodes[idx].allowMethods; ams != 0 {
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
		printSegs = append(printSegs, allNodes[i].seg)
		l := len(printSegs)
		dfs(i, allNodes, nex, printSegs)
		printSegs = printSegs[:l-1]
	}
	printSegs = []string{""}
}

// addMethodNode 添加一个节点
// path: 完整的注册路径
// allNodes: 所有节点
// nex: 节点邻接表
func addMethodToNode(method string, path string, handle http.HandlerFunc, allNodes *[]*node, nex *[][]int) {
	segs := strings.Split(path, "/")
	idx := getLastMatchedNodeIdx(segs, *allNodes, *nex)
	mCode, ok := methodMap[method]
	if !ok {
		panic("No such HTTP Method called: " + method)
	}
	lastNode := (*allNodes)[idx]
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
				(*allNodes)[(*nex)[idx][0]].seg)
		}
		idx = createNodes(idx, segs[si:], allNodes, nex)
		lastNode = (*allNodes)[idx]
	}
	lastNode.allowMethods |= mCode
	if lastNode.funcMap == nil {
		lastNode.funcMap = make(map[methodType]*http.HandlerFunc)
	}
	lastNode.funcMap[mCode] = &handle
}

// getLastMatchedNodeIdx 返回最后一个匹配的节点
func getLastMatchedNodeIdx(segs []string, allNodes []*node, nex [][]int) int {
	if len(segs) == 0 {
		return 0
	}
	que := make([]int, 1)
	root := 0
	que[0] = root
	si, idx := 1, root
	for len(que) > 0 && si < len(segs) {
		h := que[0]
		que = que[1:]
		for _, i := range nex[h] {
			if allNodes[i].seg == segs[si] {
				que = append(que, i)
				si, idx = si+1, i
				break
			}
		}
	}
	return idx
}

// createNodes 在某一节点后创建新的节点链，并且返回最后一个节点的编号
// from: 第一个不匹配的节点编号
// segs: 需要新添加的段
func createNodes(from int, segs []string, allNodes *[]*node, nex *[][]int) int {
	pre := (*allNodes)[from]
	for _, seg := range segs {
		isWild := isWildcard(seg)
		t := &node{
			seg:      seg,
			wildcard: isWild,
			level:    pre.level + 1,
		}
		ti := len(*allNodes)
		*allNodes = append(*allNodes, t)
		(*nex)[from] = append((*nex)[from], ti)
		pre.wildchild = isWild
		pre = t
		from = ti
	}
	return from
}

func isWildcard(seg string) bool {
	if len(seg) > 0 && seg[0] == ':' {
		return true
	}
	return false
}
