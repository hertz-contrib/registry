package zookeeper

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/stretchr/testify/assert"
)

func TestZookeeperRegistryAndDeregister(t *testing.T) {
	address := "127.0.0.1:8888"
	r, _ := NewZookeeperRegistry([]string{"127.0.0.1:2181"}, 40*time.Second)
	srvName := "hertz.test.demo"
	h := server.Default(
		server.WithHostPorts(address),
		server.WithRegistry(r, &registry.Info{
			ServiceName: srvName,
			Addr:        utils.NewNetAddr("tcp", address),
			Weight:      10,
			Tags:        nil,
		}))
	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"ping": "pong2"})
	})
	go h.Spin()

	time.Sleep(1 * time.Second)

	// register
	newClient, _ := client.NewClient()
	resolver, _ := NewZookeeperResolver([]string{"127.0.0.1:2181"}, 40*time.Second)
	newClient.Use(sd.Discovery(resolver))

	addr := fmt.Sprintf("http://" + srvName + "/ping")
	status, body, err := newClient.Get(context.Background(), nil, addr, config.WithSD(true))
	assert.Nil(t, err)
	assert.Equal(t, 200, status)
	assert.Equal(t, "{\"ping\":\"pong2\"}", string(body))

	// compare data
	opt := h.GetOptions()
	assert.Equal(t, opt.RegistryInfo.Weight, 10)
	assert.Equal(t, opt.RegistryInfo.Addr.String(), "127.0.0.1:8888")
	assert.Equal(t, opt.RegistryInfo.ServiceName, "hertz.test.demo")
	assert.Nil(t, opt.RegistryInfo.Tags)

	// deregister
	if err := opt.Registry.Deregister(opt.RegistryInfo); err != nil {
		t.Errorf("HERTZ: Deregister error=%v", err)
	}

	time.Sleep(5 * time.Second)

	status1, body1, err1 := newClient.Get(context.Background(), nil, addr, config.WithSD(true))
	assert.NotNil(t, err1)
	assert.Equal(t, 0, status1)
	assert.Equal(t, "", string(body1))
}
