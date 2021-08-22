package middleware

import "net/http"

// CorsMiddleware 跨域中间件
// TODO
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE")
		next.ServeHTTP(w, r)
	})
}
