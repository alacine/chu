package middleware

import (
	"log"
	"net/http"
)

// LogMiddleware 日志中间件
// TODO
func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Get new request from %v", r.Host)
		next.ServeHTTP(w, r)
	})
}
