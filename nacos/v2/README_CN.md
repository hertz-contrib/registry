# nacos (*这是一个由社区驱动的项目*)

[English](README.md)

使用 **nacos v2 sdk** 作为 **Hertz** 的注册中心

## 这个项目应当如何使用?

### 服务端

**[example/server/main.go](examples/standard/server/main.go)**

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
	r, err := nacosv2.NewDefaultNacosV2Registry()
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

### 客户端

**[example/client/main.go](examples/standard/client/main.go)**

```go
import (
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
    r, err := nacosv2.NewDefaultNacosV2Resolver()
    if err != nil {
        log.Fatal(err)
        return
    }
    client.Use(sd.Discovery(r))
	// ...
}
```

## 如何运行示例 ?

### docker 运行 nacos-server

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

```bash
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

## 自定义 Nacos Client 配置

### 服务端

**[example/custom_config/server/main.go](examples/custom_config/server/main.go)**

```go
import (
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
	
	nacosCli, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		panic(err)
	}
	
	addr := "127.0.0.1:8888"
	r := nacosRegistry.NewNacosV2Registry(nacosCli)
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

### 客户端

**[example/custom_config/client/main.go](examples/custom_config/client/main.go)**

```go
import (
	
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
    })
	
    if err != nil {
		panic(err)
    }
    r := nacos.NewNacosV2Resolver(nacosCli)
    cli.Use(sd.Discovery(r))
      // ...
} 

```

## **环境变量**

| 变量名                       | 变量默认值 | 作用 |
|---------------------------| ---------------------------------- | --------------------------------- |
| serverAddr                | 127.0.0.1                          | nacos 服务器地址 |
| serverPort                | 8848                               | nacos 服务器端口            |
| namespace                 |                                    | nacos 中的 namespace Id |

## 兼容性

Nacos 2.0 和 1.X 版本的 nacos-sdk-go 是完全兼容的，[详情](https://nacos.io/en-us/docs/2.0.0-compatibility.html)

