# registry-polaris (*This is a community driven project*)

Some application runtime use [polaris](https://github.com/polarismesh/polaris) for service discovery. Polaris is a cloud-native service discovery and governance center. 
It can be used to solve the problem of service connection, fault tolerance, traffic control and secure in distributed and microservice architecture.

## How to install registry-polaris?
```
go get -u github.com/hertz-contrib/registry/polaris
```

## How to use with Hertz server?

```go
import (
	"context"
	"log"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/registry/polaris"
)

const (
	confPath  = "polaris.yaml"
	Namespace = "Polaris"
	// At present,polaris server tag is v1.4.0，can't support auto create namespace,
	// If you want to use a namespace other than default,Polaris ,before you register an instance,
	// you should create the namespace at polaris console first.
)

func main() {
	r, err := polaris.NewPolarisRegistry(confPath)

	if err != nil {
		log.Fatal(err)
	}

	Info := &registry.Info{
		ServiceName: "hertz.test.demo",
		Addr:        utils.NewNetAddr("tcp", "127.0.0.1:8888"),
		Tags: map[string]string{
			"namespace": Namespace,
		},
	}
	h := server.Default(server.WithRegistry(r, Info), server.WithExitWaitTime(10*time.Second))

	h.GET("/hello", func(ctx context.Context, c *app.RequestContext) {
		c.String(consts.StatusOK, "Hello,Hertz!")
	})

	h.Spin()
}
```


## How to use with Hertz client?

```go
import (
	"context"
	"log"

	hclient "github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hertz-contrib/registry/polaris"
)

const (
	confPath  = "polaris.yaml"
	Namespace = "Polaris"
	// At present,polaris server tag is v1.4.0，can't support auto create namespace,
	// if you want to use a namespace other than default,Polaris ,before you register an instance,
	// you should create the namespace at polaris console first.
)

func main() {
	r, err := polaris.NewPolarisResolver(confPath)
	if err != nil {
		log.Fatal(err)
	}

	client, err := hclient.NewClient()
	client.Use(sd.Discovery(r))

	for i := 0; i < 10; i++ {
		// config.WithTag sets the namespace tag for service discovery
		status, body, err := client.Get(context.TODO(), nil, "http://hertz.test.demo/hello", config.WithSD(true), config.WithTag("namespace", Namespace))
		if err != nil {
			hlog.Fatal(err)
		}
		hlog.Infof("code=%d,body=%s\n", status, body)
	}
}
```
## How to install polaris?
Polaris support stand-alone and cluster. More information can be found in [install polaris](https://polarismesh.cn/zh/doc/%E5%BF%AB%E9%80%9F%E5%85%A5%E9%97%A8/%E5%AE%89%E8%A3%85%E6%9C%8D%E5%8A%A1%E7%AB%AF/%E5%AE%89%E8%A3%85%E5%8D%95%E6%9C%BA%E7%89%88.html#%E5%8D%95%E6%9C%BA%E7%89%88%E5%AE%89%E8%A3%85)

## Todolist
Welcome to contribute your ideas

## Use polaris with Hertz

See example

[example](example)

### Prepare

```shell
make prepare
```

### Run server

```shell
go run example/server/main.go
```

### Run client

```shell
go run example/client/main.go
```

## Compatibility

Compatible with polaris (v1.4.0 - v1.10.0), latest stable version is recommended. If you want to use other server version, please modify the version in `Makefile` to test.
