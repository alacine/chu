# chu

[![Go](https://github.com/alacine/chu/actions/workflows/go.yml/badge.svg)](https://github.com/alacine/chu/actions/workflows/go.yml)

从零开始的 Go Web 框架

- [x] 带参数的路由管理
- [x] 支持直接添加中间件
- [x] 请求 ID
- [x] 超时
- [x] 限流（普通限流、突发高并发情况限流）

TODO
- [ ] 日志
- [ ] 参数校验（功能已经实现，但是里面的校验规则只有一个样例，需要完善）
- [ ] ...

样例
```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/alacine/chu"
	"github.com/alacine/chu/middleware"
)

func main() {
	mux := chu.New()
	mux.Use(middleware.LogMiddleware)
	mux.HandleFunc(http.MethodGet, "/hello/:name", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(rw, "hello, %s\n", chu.URLParam(r, "name"))
	})
	log.Fatalln(http.ListenAndServe(":8000", mux))
}
```

server
```bash
❯ go run main.go
2021/08/22 23:06:40 Get new request from localhost:8000
```

client
```bash
❯ curl localhost:8000/hello/chu
hello, chu
```
