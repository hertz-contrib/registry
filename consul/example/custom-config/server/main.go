// Copyright 2022 CloudWeGo Authors.
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
	"log"
	"net"
	"sync"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hashicorp/consul/api"
	"github.com/hertz-contrib/registry/consul"
)

var (
	wg      sync.WaitGroup
	localIP = "your ip"
)

type Example struct {
	A int `json:"a"`
	B int `json:"b"`
}

func main() {
	config := api.DefaultConfig()
	config.Address = "127.0.0.1:8500"
	consulClient, err := api.NewClient(config)
	if err != nil {
		log.Fatal(err)
		return
	}

	// custom check
	check := &api.AgentServiceCheck{
		Interval:                       "7s",
		Timeout:                        "5s",
		DeregisterCriticalServiceAfter: "15s",
	}
	r := consul.NewConsulRegister(consulClient,
		consul.WithCheck(check),
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		addr := net.JoinHostPort(localIP, "5001")

		h := server.Default(
			server.WithHostPorts(addr),
			server.WithRegistry(r, &registry.Info{
				ServiceName: "custom-config-demo",
				Addr:        utils.NewNetAddr("tcp", addr),
				Weight:      10,
			}),
		)
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
		addr := net.JoinHostPort(localIP, "5002")
		h := server.Default(
			server.WithHostPorts(addr),
			server.WithRegistry(r, &registry.Info{
				ServiceName: "custom-config-demo",
				Addr:        utils.NewNetAddr("tcp", addr),
				Weight:      10,
			}),
		)
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
