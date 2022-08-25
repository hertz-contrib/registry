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
	scRegistry "github.com/hertz-contrib/servicecomb/registry"
)

func main() {
    const scAddr = "127.0.0.1:30100"
    const addr = "127.0.0.1:8701"
    r, err := scRegistry.NewDefaultSCRegistry([]string{scAddr})
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
	"github.com/hertz-contrib/servicecomb/resolver"
)

func main() {
    const scAddr = "127.0.0.1:30100"
	// build a servicecomb resolver 
	r, err := resolver.NewDefaultSCResolver([]string{scAddr})
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

Server：`example/server/main.go`

Client：`example/client/main.go`

## Compatibility

Compatible with consul.

maintained by: [a631807682](https://github.com/a631807682)