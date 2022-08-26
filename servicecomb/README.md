# registry-servicecomb (This is a community driven project)

Support Hertz to use ServiceComb for service registration and discovery

## Docs

### Server

#### Basic Usage

```go
import (
	"context"
	"log"
	"sync"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/servicecomb"
)

func main() {
    const scAddr = "127.0.0.1:30100"
    const addr = "127.0.0.1:8701"
    r, err := servicecomb.NewDefaultSCRegistry([]string{scAddr})
    if err != nil {
        log.Fatal(err)
        return
    }
    h := server.Default(
        server.WithHostPorts(addr),
        server.WithRegistry(r, &registry.Info{
            ServiceName: "hertz.servicecomb.demo",
            Addr:        utils.NewNetAddr("tcp", addr),
            Weight:      10,
            Tags:        nil,
        }),
    )

    h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
        ctx.JSON(consts.StatusOK, utils.H{"ping": "pong1"})
    })
    h.Spin()
}
```


### Client

```go

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hertz-contrib/servicecomb"
)

func main() {
    const scAddr = "127.0.0.1:30100"
	// build a servicecomb resolver 
	r, err := servicecomb.NewDefaultSCResolver([]string{scAddr})
	if err != nil {
		panic(err)
	}
	// build a hertz client with the servicecomb resolver
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	cli.Use(sd.Discovery(r))
	for i := 0; i < 10; i++ {
		status, body, err := cli.Get(context.Background(), nil, "http://hertz.servicecomb.demo/ping", config.WithSD(true))
		if err != nil {
			hlog.Fatal(err)
		}
		hlog.Infof("code=%d,body=%s", status, string(body))
	}
}

```

## Example

### Run Server
[Server](./example/server/main.go)
```shell
go run ./example/server/main.go
```
```log
2022/08/26 17:23:27 INFO: Use Service center v4
2022/08/26 17:23:27 INFO: Use Service center v4
2022/08/26 17:23:27.310498 engine.go:537: [Debug] HERTZ: Method=GET    absolutePath=/ping                     --> handlerName=main.main.func2.1 (num=2 handlers)
2022/08/26 17:23:27.310601 engine.go:537: [Debug] HERTZ: Method=GET    absolutePath=/ping                     --> handlerName=main.main.func1.1 (num=2 handlers)
2022/08/26 17:23:27.311129 transport.go:91: [Info] HERTZ: HTTP server listening on address=127.0.0.1:8701
2022/08/26 17:23:27.311447 transport.go:91: [Info] HERTZ: HTTP server listening on address=127.0.0.1:8702
```

### Run Client
[Client](./example/client/main.go)
```shell
go run ./example/client/main.go
```
```log
2022/08/26 17:24:03 INFO: Use Service center v4
2022/08/26 17:24:03 DEBUG: service center has new revision 9fc77257754eca927c1ff189b083e6c4eb79dbff
2022/08/26 17:24:03.413487 main.go:46: [Info] code=200,body={"ping":"pong1"}
2022/08/26 17:24:03.414199 main.go:46: [Info] code=200,body={"ping":"pong2"}
2022/08/26 17:24:03.414373 main.go:46: [Info] code=200,body={"ping":"pong1"}
2022/08/26 17:24:03.414594 main.go:46: [Info] code=200,body={"ping":"pong2"}
2022/08/26 17:24:03.414848 main.go:46: [Info] code=200,body={"ping":"pong2"}
2022/08/26 17:24:03.415051 main.go:46: [Info] code=200,body={"ping":"pong1"}
2022/08/26 17:24:03.415261 main.go:46: [Info] code=200,body={"ping":"pong2"}
2022/08/26 17:24:03.415560 main.go:46: [Info] code=200,body={"ping":"pong1"}
2022/08/26 17:24:03.415801 main.go:46: [Info] code=200,body={"ping":"pong2"}
2022/08/26 17:24:03.416111 main.go:46: [Info] code=200,body={"ping":"pong2"}
```


## Compatibility

Compatible with server (2.0.0 - latest), If you want to use older server version, please modify the version in `Makefile` to test.

maintained by: [Cr](https://github.com/a631807682)