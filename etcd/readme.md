# etcd (*This is a community driven project*)

Etcd as service discovery for Hertz.

Your etcd version must be v3
## how to use?

### server 

**[example/server/main.go](example/server/main.go)**

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
	"go.etcd.io/etcd/clientv3"
)

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 2 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	r, err := etcd.NewEtcdRegistry(cli, time.Second*2)
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
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"go.etcd.io/etcd/clientv3"

	"github.com/hertz-contrib/registry/etcd"
)

var wg sync.WaitGroup

func main() {

	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 2 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	r, err := etcd.NewEtcdResolver(etcdCli, 2*time.Second)
	if err != nil {
		panic(err)
	}
	cli.Use(sd.Discovery(r))
	for i := 0; i < 10; i++ {
		status, body, err := cli.Get(context.Background(), nil, "http://hertz.test.demo/ping", config.WithSD(true))
		if err != nil {
			hlog.Fatal(err)
		}
		hlog.Infof("code=%d,body=%s", status, string(body))
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