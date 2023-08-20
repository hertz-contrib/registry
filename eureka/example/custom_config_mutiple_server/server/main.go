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
	"net"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/registry/eureka"
	"github.com/hudl/fargo"
)

var (
	wg         sync.WaitGroup
	configPath = "paht/to/your/config/file.gcfg"
)

type Message struct {
	Message string `json:"message"`
}

func main() {
	// custom config
	eurekaConfig, err := fargo.ReadConfig(configPath)
	if err != nil {
		hlog.Fatal(err)
	}
	r := eureka.NewEurekaRegistryFromConfig(eurekaConfig, 40*time.Second)

	wg.Add(2)
	go func() {
		defer wg.Done()
		addr := net.JoinHostPort("127.0.0.1", "5001")

		h := server.Default(
			server.WithHostPorts(addr),
			server.WithRegistry(r, &registry.Info{
				ServiceName: "hertz.discovery.eureka",
				Addr:        utils.NewNetAddr("tcp", addr),
				Weight:      10,
				Tags: map[string]string{
					"key1": "val1",
				},
			}),
		)
		h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
			ctx.JSON(consts.StatusOK, utils.H{"ping": "pong"})
		})

		h.POST("ping", func(c context.Context, ctx *app.RequestContext) {
			m := Message{}
			if err := ctx.Bind(&m); err != nil {
				ctx.String(consts.StatusBadRequest, err.Error())
				return
			}
			ctx.JSON(consts.StatusOK, m)
		})
		h.Spin()
	}()

	go func() {
		defer wg.Done()
		addr := net.JoinHostPort("127.0.0.1", "5002")

		h := server.Default(
			server.WithHostPorts(addr),
			server.WithRegistry(r, &registry.Info{
				ServiceName: "hertz.discovery.eureka",
				Addr:        utils.NewNetAddr("tcp", addr),
				Weight:      10,
				Tags: map[string]string{
					"key1": "val1",
				},
			}),
		)
		h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
			ctx.JSON(consts.StatusOK, utils.H{"ping": "pong"})
		})

		h.POST("ping", func(c context.Context, ctx *app.RequestContext) {
			m := Message{}
			if err := ctx.Bind(&m); err != nil {
				ctx.String(consts.StatusBadRequest, err.Error())
				return
			}
			ctx.JSON(consts.StatusOK, m)
		})
		h.Spin()
	}()

	wg.Wait()
}
