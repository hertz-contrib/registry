// Copyright 2021 CloudWeGo Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"

	"sync"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/registry/nacos"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

var (
	wg        sync.WaitGroup
	server1IP = "127.0.0.1:8088"
	server2IP = "127.0.0.1:8089"
)

type Message struct {
	Message string `json:"message"`
	Name    string `json:"name"`
}

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
	wg.Add(2)
	r := nacos.NewNacosRegistry(cli)
	go func() {
		defer wg.Done()
		h := server.Default(
			server.WithHostPorts(server1IP),
			server.WithRegistry(r, &registry.Info{
				ServiceName: "hertz.custom-config.demo",
				Addr:        utils.NewNetAddr("tcp", server1IP),
				Weight:      10,
				Tags: map[string]string{
					"key1": "val1",
				},
			}))

		h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
			ctx.JSON(consts.StatusOK, utils.H{"ping1": "pong1"})
		})
		h.POST("/hello", func(c context.Context, ctx *app.RequestContext) {
			message := Message{}
			if err := ctx.Bind(&message); err != nil {
				ctx.String(consts.StatusBadRequest, err.Error())
				return
			}
			ctx.JSON(consts.StatusOK, message)
		})

		h.Spin()
	}()

	go func() {
		defer wg.Done()
		h := server.Default(
			server.WithHostPorts(server2IP),
			server.WithRegistry(r, &registry.Info{
				ServiceName: "hertz.custom-config.demo",
				Addr:        utils.NewNetAddr("tcp", server2IP),
				Weight:      10,
				Tags: map[string]string{
					"key2": "val2",
				},
			}))
		h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
			ctx.JSON(consts.StatusOK, utils.H{"ping2": "pong2"})
		})

		h.POST("/hello", func(c context.Context, ctx *app.RequestContext) {
			message := Message{}
			if err := ctx.Bind(&message); err != nil {
				ctx.String(consts.StatusBadRequest, err.Error())
				return
			}
			ctx.JSON(consts.StatusOK, message)
		})
		h.Spin()
	}()
	wg.Wait()
}
