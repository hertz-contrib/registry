# registry-eureka (*This is a community driven project*)

registry-eureka implements Hertz registry and resolver for Netflix Eureka. 

## How to use?

### Server

**[example/server/main.go](examples/server/main.go)**

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
	"github.com/hertz-contrib/registry/eureka"
)

func main() {
	addr := "127.0.0.1:8888"
	r := eureka.NewEurekaRegistry([]string{"http://127.0.0.1:8761/eureka"}, 40*time.Second)
	h := server.Default(
		server.WithHostPorts(addr),
		server.WithRegistry(r, &registry.Info{
			ServiceName: "hertz.discovery.eureka",
			Addr:        utils.NewNetAddr("tcp", addr),
			Weight:      10,
			Tags:        nil,
		}))
	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"ping": "pong2"})
	})
	h.Spin()
}

```

### Client

**[example/client/main.go](example/basic/client/main.go)**

```go
package main

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hertz-contrib/registry/eureka"
)

func main() {
	cli, err := client.NewClient()
	if err != nil {
		hlog.Fatal(err)
		return
	}
	r := eureka.NewEurekaResolver([]string{"http://127.0.0.1:8761/eureka"})

	cli.Use(sd.Discovery(r))
	for i := 0; i < 10; i++ {
		status, body, err := cli.Get(context.Background(), nil, "http://hertz.discovery.eureka/ping", config.WithSD(true))
		if err != nil {
			hlog.Fatal(err)
		}
		hlog.Infof("code=%d,body=%s", status, string(body))
	}
}

```

## How to Run example?

### Start Eureka Server

```bash
docker-compose up
```

### Run Server

```go
go run ./example/server/main.go
```

### Run Client

```go
go run ./example/client/main.go
```

```go
2022/08/28 22:16:59 Getting app hertz.discovery.eureka from url http://127.0.0.1:8761/eureka/apps/hertz.discovery.eureka
2022/08/28 22:16:59 Got eureka response from url=http://127.0.0.1:8761/eureka/apps/hertz.discovery.eureka
2022/08/28 22:16:59.443078 main.go:41: [Info] code=200,body={"ping":"pong2"}
2022/08/28 22:16:59.443258 main.go:41: [Info] code=200,body={"ping":"pong2"}
2022/08/28 22:16:59.443405 main.go:41: [Info] code=200,body={"ping":"pong2"}
2022/08/28 22:16:59.443548 main.go:41: [Info] code=200,body={"ping":"pong2"}
2022/08/28 22:16:59.443697 main.go:41: [Info] code=200,body={"ping":"pong2"}
2022/08/28 22:16:59.443855 main.go:41: [Info] code=200,body={"ping":"pong2"}
2022/08/28 22:16:59.444004 main.go:41: [Info] code=200,body={"ping":"pong2"}
2022/08/28 22:16:59.444149 main.go:41: [Info] code=200,body={"ping":"pong2"}
2022/08/28 22:16:59.444289 main.go:41: [Info] code=200,body={"ping":"pong2"}
2022/08/28 22:16:59.444405 main.go:41: [Info] code=200,body={"ping":"pong2"}


```

## Configuration

This project uses [fargo](https://github.com/hudl/fargo) as eureka client. You should refer to
[fargo](https://github.com/hudl/fargo) documentation for advanced configuration. 


There are multiple ways to crate a `eurekaRegistry`. 
- `NewEurekaRegistry`  creates a registry with a slice of eureka server addresses.
- `NewEurekaRegistryFromConfig` creates a registry with given `fargo.Config`.
- `NewEurekaRegistryFromConn` creates a registry using existing `fargo.EurekaConnection` .

The same also applies for `eurekaResolver`.
- `NewEurekaResolver`  creates a resolver with a slice of eureka server addresses.
- `NewEurekaResolverFromConfig`  creates a resolver with given `fargo.Config`.
- `NewEurekaResolverFromConn` creates a resolver using existing `fargo.EurekaConnection` .

### Authentication
A straight-forward approach is passing [credentials in uri](https://developer.mozilla.org/en-US/docs/Web/HTTP/Authentication#access_using_credentials_in_the_url) e.g. `[]string{"http://username:password@127.0.0.1:8080/eureka"`.
Alternatively, you can pass existing connection to `NewEurekaRegistryFromConn` or `NewEurekaResolverFromConn`.

### Setting Log Level

As discussed above, this project uses fargo as eureka client, which relies on [go-logging](github.com/op/go-logging) for logging.
Unfortunately, [go-logging](github.com/op/go-logging) does not provide an interface to adjust log level. The following code demonstrates how to set log level.
```go

package main

import (
	"context"
	"github.com/op/go-logging"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/registry/eureka"
)

func main() {

	logging.SetLevel(logging.WARNING, "fargo")
	// set this to a higher level if you wish to check responses from eureka 
	logging.SetLevel(logging.WARNING, "fargo.metadata")
	logging.SetLevel(logging.WARNING, "fargo.marshal")

	addr := "127.0.0.1:8888"
	r := eureka.NewEurekaRegistry([]string{"http://127.0.0.1:8761/eureka"}, 40*time.Second)
	h := server.Default(
		server.WithHostPorts(addr),
		server.WithRegistry(r, &registry.Info{
			ServiceName: "hertz.discovery.eureka",
			Addr:        utils.NewNetAddr("tcp", addr),
			Weight:      10,
			Tags:        nil,
		}))
	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"ping": "pong2"})
	})
	h.Spin()
}


```




## Compatibility

This project is compatible with eureka server v1.