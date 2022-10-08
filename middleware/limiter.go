package middleware

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/alacine/chu"
)

var curConn int64

// Limiter 限流中间件（限制同时可以处理的请求数量）
func Limiter(limit int) chu.Middleware {
	curConn = 0
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if curConn > int64(limit) {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			atomic.AddInt64(&curConn, 1)
			defer func() {
				atomic.AddInt64(&curConn, -1)
			}()
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// BurstBucketLimiter 应对突发高频率请求的限流
// limit: 突发最大并发数量
// interval: bucket 每填充一个 token 的时间间隔
func BurstBucketLimiter(limit int, interval time.Duration) chu.Middleware {
	bucket := make(chan struct{}, limit)
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			bucket <- struct{}{}
			//log.Printf("Put a new token")
		}
	}()
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-bucket:
				next.ServeHTTP(w, r)
			default:
				w.WriteHeader(http.StatusServiceUnavailable)
			}
		}
		return http.HandlerFunc(fn)
	}
}
