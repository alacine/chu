package middleware

import (
	"log"
	"net/http"
)

// LogMiddleware 日志中间件
// TODO
func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := GetRequestID(r.Context())
		log.Printf("%s %s %s", reqID, r.Host, r.URL)
		next.ServeHTTP(w, r)
	})
}
