# nacos (*这是一个由社区驱动的项目*)

[English](README.md)

使用 **nacos** 作为 **Hertz** 的注册中心

## 这个项目应当如何使用?

### 注意

- nacos/v2 版本中 hertz 目前不支持多次在同分组下创建多端口示例
- nacos/v2 的服务注册与发现和先前的版本兼容
- nacos-sdk-go v2 版本中 constant.ClientConfig 中 CustomLogger 类型被移除, 转而使用 (github.com/nacos-group/nacos-sdk-go/v2/common/logger).SetLogger 方法进行日志自定义
- nacos/v2 只支持 nacos 2.X 版本

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
    "github.com/hertz-contrib/registry/nacos/v2"
)

func main() {
	addr := "127.0.0.1:8888"
	r, err := nacosRegistry.NewDefaultNacosRegistry()
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
	// ...
	h.Spin()
}

```

### 客户端

**[example/client/main.go](examples/standard/client/main.go)**

```go
import (
    "context"
    "log"
    
    "github.com/cloudwego/hertz/pkg/app/client"
    "github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
    "github.com/cloudwego/hertz/pkg/common/config"
    "github.com/cloudwego/hertz/pkg/common/hlog"
    "github.com/hertz-contrib/registry/nacos/v2"
)

func main() {
	client, err := client.NewClient()
	if err != nil {
		panic(err)
	}
    r, err := nacos.NewDefaultNacosResolver()
    if err != nil {
        log.Fatal(err)
        return
    }
    client.Use(sd.Discovery(r))
	// ...
}
```

### 自定义日志

**[examples/logger/main.go](examples/logger/main.go)**

```go
package main

import (
	"github.com/hertz-contrib/registry/nacos/v2/common"
	"github.com/nacos-group/nacos-sdk-go/v2/common/logger"
)

func main() {
	logger.InitLogger(logger.Config{
		Level: "debug",
	})
	logger.SetLogger(common.NewCustomNacosLogger())
	logger.Info("info")
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
2022/07/26 13:52:47.310617 main.go:46: [Info] code = 200, body ={"ping":"pong"}
2022/07/26 13:52:47.311019 main.go:46: [Info] code = 200, body ={"ping":"pong"}
2022/07/26 13:52:47.311186 main.go:46: [Info] code = 200, body ={"ping":"pong"}
2022/07/26 13:52:47.311318 main.go:46: [Info] code = 200, body ={"ping":"pong"}
2022/07/26 13:52:47.311445 main.go:46: [Info] code = 200, body ={"ping":"pong"}
2022/07/26 13:52:47.311585 main.go:46: [Info] code = 200, body ={"ping":"pong"}
2022/07/26 13:52:47.311728 main.go:46: [Info] code = 200, body ={"ping":"pong"}
2022/07/26 13:52:47.311858 main.go:46: [Info] code = 200, body ={"ping":"pong"}
2022/07/26 13:52:47.311977 main.go:46: [Info] code = 200, body ={"ping":"pong"}
2022/07/26 13:52:47.312107 main.go:46: [Info] code = 200, body ={"ping":"pong"}
```

## 自定义 Nacos Client 配置

### 服务端

**[example/custom_config/server/main.go](examples/custom_config/server/main.go)**

```go
import (
    "context"
    
    "github.com/cloudwego/hertz/pkg/app"
    "github.com/cloudwego/hertz/pkg/app/server"
    "github.com/cloudwego/hertz/pkg/app/server/registry"
    "github.com/cloudwego/hertz/pkg/common/utils"
    "github.com/cloudwego/hertz/pkg/protocol/consts"
    "github.com/hertz-contrib/registry/nacos/v2"
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
	r := nacosRegistry.NewNacosRegistry(nacosCli)
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
    "context"
    
    "github.com/cloudwego/hertz/pkg/app/client"
    "github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
    "github.com/cloudwego/hertz/pkg/common/config"
    "github.com/cloudwego/hertz/pkg/common/hlog"
    "github.com/hertz-contrib/registry/nacos/v2"
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
    r := nacos.NewNacosResolver(nacosCli)
    cli.Use(sd.Discovery(r))
      // ...
} 

```

## **环境变量**

| 变量名 | 变量默认值 | 作用 |
| ------------------------- | ---------------------------------- | --------------------------------- |
| serverAddr               | 127.0.0.1                          | nacos 服务器地址 |
| serverPort               | 8848                               | nacos 服务器端口            |
| namespace                 |                                    | nacos 中的 namespace Id |

## 兼容性

- 使用 Nacos2.x 客户端

- Nacos 2.x [详细信息](https://nacos.io/en-us/docs/v2/upgrading/2.0.0-compatibility.html)

- Go 版本不低于 1.16

- 支持Nacos 2.x

