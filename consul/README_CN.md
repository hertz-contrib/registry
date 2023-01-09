# registry-consul (这是一个由社区驱动的项目)

[English](README.md)

支持**Hertz**使用**Consul**进行服务注册与发现

## 文档

### 服务端

#### 基本使用

```golang
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/hertz-contrib/registry/consul"
)

func main() {
	// build a consul client
	config := consulapi.DefaultConfig()
	config.Address = "127.0.0.1:8500"
	consulClient, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal(err)
		return
	}
	// build a consul register with the consul client
	r := consul.NewConsulRegister(consulClient)

	// run Hertz with the consul register
	localIP, err := consul.GetLocalIPv4Address()
	if err != nil {
		log.Fatal(err)
	}
	addr := fmt.Sprintf("%s:8888",localIP)
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
		ctx.JSON(consts.StatusOK, utils.H{"ping": "pong1"})
	})
	h.Spin()
}
```

#### 自定义服务检查

注册中心默认配置服务检查，如下：

```
check.Timeout = "5s"
check.Interval = "5s"
check.DeregisterCriticalServiceAfter = "1m"
```

你也可以使用`WithCheck`来修改配置

```golang
package main

import (
	"log"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/hertz-contrib/registry/consul"
)

func main() {
	// build a consul client
	config := consulapi.DefaultConfig()
	config.Address = "127.0.0.1:8500"
	consulClient, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal(err)
		return
	}

	// build a consul register with the check option
	check := new(consulapi.AgentServiceCheck)
	check.Timeout = "10s"
	check.Interval = "10s"
	check.DeregisterCriticalServiceAfter = "1m"
	r := consul.NewConsulRegister(consulClient, consul.WithCheck(check))
}

```

### 客户端

```golang
package main

import (
	"log"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/hertz-contrib/registry/consul"
)

func main() {
	// build a consul client
	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = "127.0.0.1:8500"
	consulClient, err := consulapi.NewClient(consulConfig)
	if err != nil {
		log.Fatal(err)
		return
	}
	// build a consul resolver with the consul client
	r := consul.NewConsulResolver(consulClient)

	// build a hertz client with the consul resolver
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	cli.Use(sd.Discovery(r))
}
```

## 使用样例

[服务端](example/basic/server/main.go)：`example/server/main.go`

[客户端](example/basic/client/main.go)：`example/client/main.go`

## 兼容性

与Consul(**v1.11.x到v1.13.x**)保持兼容。

[consul版本列表](https://releases.hashicorp.com/consul)

维护者: [Lemonfish](https://github.com/LemonFish873310466) / [claude-zq](https://github.com/Claude-Zq)