package chu

import (
	"net/http"
)

// Params URL 中参数的存放方式，key 和 value 序号一一对应
type Params struct {
	Keys, Values []string
}

// Context chu 的 Context，目前只存放 URP 参数
type Context struct {

	// URL path 中的参数
	URLParams Params
}

// URLParam 从 http.Request 中或取 URL 参数
func URLParam(r *http.Request, name string) string {
	if ctx, _ := r.Context().Value(ContextKey).(*Context); ctx != nil {
		return ctx.URLParam(name)
	}
	return ""
}

type contextKey string

// ContextKey ...
var ContextKey = contextKey("ChuContextKey")

// NewChuContext return a *Context
func NewChuContext() *Context {
	return &Context{}
}

// Reset ...
func (c *Context) Reset() {
	c.URLParams.Keys = c.URLParams.Keys[:0]
	c.URLParams.Values = c.URLParams.Values[:0]
	//c.parentCtx = nil
}

// URLParam 获取 Context.URLParams 中对应 Param
func (c *Context) URLParam(name string) string {
	for i := 0; i < len(c.URLParams.Keys); i++ {
		if c.URLParams.Keys[i] == name {
			return c.URLParams.Values[i]
		}
	}
	return ""
}
