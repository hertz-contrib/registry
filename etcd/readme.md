# etcd (*This is a community driven project*)

Etcd as service discovery for Hertz.

## How to use?

### Server 

**[example/server/main.go](example/server/main.go)**

```go
package main

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/registry/etcd"
)

func main() {
	r, err := etcd.NewEtcdRegistry([]string{"127.0.0.1:2379"})
	if err != nil {
		panic(err)
	}
	addr := "127.0.0.1:8888"
	h := server.Default(
		server.WithHostPorts(addr),
		server.WithRegistry(r, &registry.Info{
			ServiceName: "hertz.test.demo",
			Addr:        utils.NewNetAddr("tcp", addr),
			Weight:      10,
			Tags:        nil,
		}))
	h.GET("/ping", func(_ context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"ping": "pong2"})
	})
	h.Spin()
}
```

### Client

**[example/client/main.go](example/client/main.go)**

```go
package main

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hertz-contrib/registry/etcd"
)

func main() {
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	r, err := etcd.NewEtcdResolver([]string{"127.0.0.1:2379"})
	if err != nil {
		panic(err)
	}
	cli.Use(sd.Discovery(r))
	for i := 0; i < 10; i++ {
		status, body, err := cli.Get(context.Background(), nil, "http://hertz.test.demo/ping", config.WithSD(true))
		if err != nil {
			hlog.Fatal(err)
		}
		hlog.Infof("HERTZ: code=%d,body=%s", status, string(body))
	}
}
```
## How to run example?

### run docker

```bash
make prepare
```

### Run etcd cluster

```shell
make prepare-cluster
```

### run server

```go
go run ./example/server/main.go
```

### run client

```go
go run ./example/client/main.go
```
```go
/hertz.test.demo/127.0.0.1:8888:{"Weight":10,"Tags":null}
2022/08/23 21:11:29.109063 main.go:57: [Info] code=200,body={"ping":"pong2"}
2022/08/23 21:11:29.109268 main.go:57: [Info] code=200,body={"ping":"pong2"}
2022/08/23 21:11:29.109377 main.go:57: [Info] code=200,body={"ping":"pong2"}
2022/08/23 21:11:29.109523 main.go:57: [Info] code=200,body={"ping":"pong2"}
2022/08/23 21:11:29.109887 main.go:57: [Info] code=200,body={"ping":"pong2"}
2022/08/23 21:11:29.112675 main.go:57: [Info] code=200,body={"ping":"pong2"}
2022/08/23 21:11:29.113081 main.go:57: [Info] code=200,body={"ping":"pong2"}
2022/08/23 21:11:29.114662 main.go:57: [Info] code=200,body={"ping":"pong2"}
2022/08/23 21:11:29.114854 main.go:57: [Info] code=200,body={"ping":"pong2"}
2022/08/23 21:11:29.115257 main.go:57: [Info] code=200,body={"ping":"pong2"}
```

## Authentication

### Server

```go
package main

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/registry/etcd"
)

func main() {
	r, err := etcd.NewEtcdRegistry([]string{"127.0.0.1:2379"}, etcd.WithAuthOpt("root", "123456"))
	if err != nil {
		panic(err)
	}
	addr := "127.0.0.1:8888"
	h := server.Default(
		server.WithHostPorts(addr),
		server.WithRegistry(r, &registry.Info{
			ServiceName: "hertz.test.demo",
			Addr:        utils.NewNetAddr("tcp", addr),
			Weight:      10,
			Tags:        nil,
		}))
	h.GET("/ping", func(_ context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"ping": "pong2"})
	})
	h.Spin()
}
```

### Client

```go
package main

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hertz-contrib/registry/etcd"
)

func main() {
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	r, err := etcd.NewEtcdResolver([]string{"127.0.0.1:2379"}, etcd.WithAuthOpt("root", "123456"))
	if err != nil {
		panic(err)
	}
	cli.Use(sd.Discovery(r))
	for i := 0; i < 10; i++ {
		status, body, err := cli.Get(context.Background(), nil, "http://hertz.test.demo/ping", config.WithSD(true))
		if err != nil {
			hlog.Fatal(err)
		}
		hlog.Infof("HERTZ: code=%d,body=%s", status, string(body))
	}
}
```
## Retry

After the service is registered to `ETCD`, it will regularly check the status of the service. If any abnormal status is found, it will try to register the service again. `observeDelay` is the delay time for checking the service status under normal conditions, and `retryDelay` is the delay time for attempting to register the service after disconnecting.

### Default Retry Config

| Config Name         | Default Value    | Description                                                                               |
|:--------------------|:-----------------|:------------------------------------------------------------------------------------------|
| WithMaxAttemptTimes | 5                | Used to set the maximum number of attempts, if 0, it means infinite attempts              |
| WithObserveDelay    | 30 * time.Second | Used to set the delay time for checking service status under normal connection conditions |
| WithRetryDelay      | 10 * time.Second | Used to set the retry delay time after disconnecting                                      |

### Example

```go
package main

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/registry/etcd"
)

func main() {
	r, _ := etcd.NewEtcdRegistry(
		[]string{"127.0.0.1:2379"},
		etcd.WithMaxAttemptTimes(10),
		etcd.WithObserveDelay(20*time.Second),
		etcd.WithRetryDelay(5*time.Second),
	)

	addr := "127.0.0.1:8888"
	h := server.Default(
		server.WithHostPorts(addr),
		server.WithRegistry(r, &registry.Info{
			ServiceName: "hertz.test.demo",
			Addr:        utils.NewNetAddr("tcp", addr),
			Weight:      10,
			Tags:        nil,
		}),
	)
	h.GET("/ping", func(_ context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"ping": "pong2"})
	})
	h.Spin()
}

```

## How to Dynamically specify ip and port

To dynamically specify an IP and port, one should first set the environment variables `HERTZ_IP_TO_REGISTRY` and `HERTZ_PORT_TO_REGISTRY`. If these variables are not set, the system defaults to using the service's listening IP and port. Notably, if the service's listening IP is either not set or set to "::", the system will automatically retrieve and use the machine's IPV4 address.

## Compatibility

Compatible with server (3.0.0 - 3.5.4)
etcd-clientv3 [see](https://github.com/etcd-io/etcd/tree/main/client/v3)
