# chu

从零开始的 Go Web 框架

- [x] 带参数的路由管理
- [x] 支持直接添加中间件

TODO
- [ ] 日志
- [ ] 参数校验
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
