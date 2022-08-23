# zookeeper (*This is a community driven project*)

Zookeeper as service discovery for Hertz.

## how to use?

### server 

**[example/basic/server/main.go](example/basic/server/main.go)**

```go
import (
    "context"
    "time"

    "github.com/cloudwego/hertz/pkg/app"
    "github.com/cloudwego/hertz/pkg/app/server"
    "github.com/cloudwego/hertz/pkg/app/server/registry"
    "github.com/cloudwego/hertz/pkg/common/utils"
    "github.com/cloudwego/hertz/pkg/protocol/consts"

    "github.com/hertz-contrib/registry/zookeeper"
)

func main() {
    addr := "127.0.0.1:8888"
    r, err := zookeeper.NewZookeeperRegistry([]string{"127.0.0.1:2181"}, 40*time.Second)
    if err != nil {
        panic(err)
    }
    h := server.Default(
        server.WithHostPorts(addr),
        server.WithRegistry(r, &registry.Info{
            ServiceName: "hertz.test.demo",
            Addr:        utils.NewNetAddr("tcp", addr),
            Weight:      10,
            Tags:        nil,
        }))
    h.GET("/ping", func(c context.Context, ctx *app.RequestContext) 		{
        ctx.JSON(consts.StatusOK, utils.H{"ping": "pong2"})
    })
    h.Spin()
}

```

### Client

**[example/basic/client/main.go](example/basic/client/main.go)**

```go
import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	"github.com/hertz-contrib/registry/zookeeper"
)

func main() {
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	r, err :=zookeeper.NewZookeeperResolver([]string{"127.0.0.1:2181"}, 40*time.Second)
	if err != nil {
		panic(err)
	}
	cli.Use(sd.Discovery(r))
// ...
}
```
## How to run example?

### run docker

- make prepare

```bash
make prepare
```

### run server

```go
go run ./example/basic/server/main.go
```

### run client

```go
go run ./example/basic/client/main.go
```
```go
2022/08/21 23:31:59.391243 main.go:44: [Info] code=200,body={"ping":"pong2"}
2022/08/21 23:31:59.391493 main.go:44: [Info] code=200,body={"ping":"pong2"}
2022/08/21 23:31:59.391603 main.go:44: [Info] code=200,body={"ping":"pong2"}
2022/08/21 23:31:59.391714 main.go:44: [Info] code=200,body={"ping":"pong2"}
2022/08/21 23:31:59.391816 main.go:44: [Info] code=200,body={"ping":"pong2"}
2022/08/21 23:31:59.391913 main.go:44: [Info] code=200,body={"ping":"pong2"}
2022/08/21 23:31:59.392039 main.go:44: [Info] code=200,body={"ping":"pong2"}
2022/08/21 23:31:59.392144 main.go:44: [Info] code=200,body={"ping":"pong2"}
2022/08/21 23:31:59.392249 main.go:44: [Info] code=200,body={"ping":"pong2"}
2022/08/21 23:31:59.392379 main.go:44: [Info] code=200,body={"ping":"pong2"}
```