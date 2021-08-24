package middleware

//https://github.com/zenazn/goji/tree/master/web/middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
)

type requestIDCtxKey string

// RequestIDKey 为 RequestID 在 Context 中的 Key
const RequestIDKey requestIDCtxKey = "requestIDCtxKey"

var RequestIDHeader = "X-Request-Id"
var prefix string
var reqid uint64

func init() {
	hostname, err := os.Hostname()
	if hostname == "" || err != nil {
		hostname = "localhost"
	}
	prefix = fmt.Sprintf("%s-", hostname)
}

// RequestID 记录请求 ID
func RequestID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		id := atomic.AddUint64(&reqid, 1)
		reqId := fmt.Sprintf("%s%d", prefix, id)
		r = r.WithContext(context.WithValue(r.Context(), RequestIDKey, reqId))
		w.Header().Add(RequestIDHeader, reqId)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// GetRequestID 从 Context 中获取 RequestID
func GetRequestID(c context.Context) string {
	id, ok := c.Value(RequestIDKey).(string)
	if !ok {
		return ""
	}
	return id
}
