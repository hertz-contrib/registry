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
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/registry/zookeeper"
	"net"
	"sync"
	"time"
)

var wg sync.WaitGroup

type Example struct {
	A int `json:"a"`
	B int `json:"b"`
}

func main() {
	r, err := zookeeper.NewZookeeperRegistry([]string{"127.0.0.1:2181"}, 40*time.Second)
	if err != nil {
		panic(err)
	}

	wg.Add(2)
	go func() {
		defer wg.Done()
		addr := net.JoinHostPort("127.0.0.1", "8888")
		tags := map[string]string{"group": "blue", "idc": "hd1"}
		h := server.Default(
			server.WithHostPorts(addr),
			server.WithRegistry(r, &registry.Info{
				ServiceName: "hertz.test.demo1",
				Addr:        utils.NewNetAddr("tcp", addr),
				Weight:      10,
				Tags:        tags,
			}))
		h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
			ctx.JSON(consts.StatusOK, utils.H{"ping": "pong1"})
		})
		h.POST("/ping", func(c context.Context, ctx *app.RequestContext) {
			e := Example{}
			if err := ctx.Bind(&e); err != nil {
				ctx.String(consts.StatusBadRequest, err.Error())
				return
			}
			ctx.JSON(consts.StatusOK, e)
		})
		h.Spin()

	}()
	go func() {
		defer wg.Done()
		addr := net.JoinHostPort("127.0.0.1", "8889")
		tags := map[string]string{"group": "red", "idc": "hd2"}
		h := server.Default(
			server.WithHostPorts(addr),
			server.WithRegistry(r, &registry.Info{
				ServiceName: "hertz.test.demo2",
				Addr:        utils.NewNetAddr("tcp", addr),
				Weight:      10,
				Tags:        tags,
			}))
		h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
			ctx.JSON(consts.StatusOK, utils.H{"ping": "pong2"})
		})
		h.POST("/ping", func(c context.Context, ctx *app.RequestContext) {
			e := Example{}
			if err := ctx.Bind(&e); err != nil {
				ctx.String(consts.StatusBadRequest, err.Error())
				return
			}
			ctx.JSON(consts.StatusOK, e)
		})
		h.Spin()

	}()

	wg.Wait()
}
