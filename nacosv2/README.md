# nacos (*This is a community driven project*)

[中文](../nacos/README_CN.md)

NacosV2 as service discovery for Hertz.

## How to use?

### Server

**[example/standard/server/main.go](../nacos/examples/standard/server/main.go)**

```go
import (
	"context"
	"log"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/registry/nacosv2"
)

func main() {
	addr := "127.0.0.1:8888"
	r, err := nacos.NewDefaultNacosV2Registry()
	if err != nil {
		log.Fatal(err)
		return
	}
	h := server.Default(
		server.WithHostPorts(addr),
		server.WithRegistry(r, &registry.Info{
			ServiceName: "hertz.test.demo",
			Addr:        utils.NewNetAddr("tcp", addr),
			Weight:      10,
			Tags:        nil,
		}),
	)
	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"ping": "pong"})
	})
	h.Spin()
}
```

### Client

**[example/standard/client/main.go](../nacos/examples/standard/client/main.go)**

```go
import (
	"context"
	"log"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hertz-contrib/registry/nacosv2"
)

func main() {
client, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	r, err := nacos.NewDefaultNacosV2Resolver()
	if err != nil {
		log.Fatal(err)
		return
	}
	client.Use(sd.Discovery(r))
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
go run ./examples/standard/server/main.go
```

### run client

```go
go run ./examples/standard/client/main.go
```

```go
2024/01/04 15:30:12.841911 main.go:81: [Info] code=200,body={"ping2":"pong2"}
2024/01/04 15:30:12.842074 main.go:81: [Info] code=200,body={"ping2":"pong2"}
2024/01/04 15:30:12.842207 main.go:81: [Info] code=200,body={"ping2":"pong2"}
2024/01/04 15:30:12.842372 main.go:81: [Info] code=200,body={"ping2":"pong2"}
2024/01/04 15:30:12.842513 main.go:81: [Info] code=200,body={"ping2":"pong2"}
2024/01/04 15:30:12.842637 main.go:81: [Info] code=200,body={"ping2":"pong2"}
2024/01/04 15:30:12.842765 main.go:81: [Info] code=200,body={"ping2":"pong2"}
2024/01/04 15:30:12.842876 main.go:81: [Info] code=200,body={"ping2":"pong2"}
2024/01/04 15:30:12.842965 main.go:81: [Info] code=200,body={"ping2":"pong2"}
2024/01/04 15:30:12.843047 main.go:81: [Info] code=200,body={"ping2":"pong2"}
```

## Custom Nacos Client Configuration

### Server

**[example/custom_config/server/main.go](../nacos/examples/custom_config/server/main.go)**

```go
import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/registry/nacosv2"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

func main() {
	sc := []constant.ServerConfig{
		*constant.NewServerConfig("127.0.0.1", 8848),
	}
	
	cc := constant.ClientConfig{
		NamespaceId:         "public",
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "info",
	}
	
	cli, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		panic(err)
	}
	
	addr := "127.0.0.1:8888"
	r := nacos.NewNacosV2Registry(cli)
	h := server.Default(
		server.WithHostPorts(addr),
		server.WithRegistry(r, &registry.Info{
			ServiceName: "hertz.test.demo",
			Addr:        utils.NewNetAddr("tcp", addr),
			Weight:      10,
			Tags:        nil,
		}),
	)
	// ...
	h.Spin()
}

```

### Client

**[example/custom_config/client/main.go](../nacos/examples/custom_config/client/main.go)**

```go
import (
	"context"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hertz-contrib/registry/nacosv2"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

func main() {
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	sc := []constant.ServerConfig{
		*constant.NewServerConfig("127.0.0.1", 8848),
	}
	cc := constant.ClientConfig{
		NamespaceId:         "public",
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "info",
	}

	nacosCli, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		}
  )
	if err != nil {
		panic(err)
	}
	r := nacos.NewNacosV2Resolver(nacosCli)
	cli.Use(sd.Discovery(r))
	// ...
}

```

## Environment Variable

| Environment Variable Name | Environment Variable Default Value | Environment Variable Introduction |
|---------------------------| ---------------------------------- | --------------------------------- |
| serverAddr                | 127.0.0.1                          | nacos server address              |
| serverPort                | 8848                               | nacos server port                 |
| namespace                 |                                    | the namespaceId of nacos          |

## Compatibility

The server of Nacos2.0 is fully compatible with 1.X
nacos-sdk-go. [see](https://nacos.io/en-us/docs/2.0.0-compatibility.html)
