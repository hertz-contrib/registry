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
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/go-chassis/sc-client"
	"github.com/hertz-contrib/registry/servicecomb"
)

var wg sync.WaitGroup

type Example struct {
	A int `json:"a"`
	B int `json:"b"`
}

const scAddr = "127.0.0.1:30100"

func main() {
	// custom config
	scClient, err := sc.NewClient(sc.Options{
		Endpoints:  []string{scAddr},
		Timeout:    5 * time.Second,
		Compressed: true,
	})
	if err != nil {
		panic(err)
	}

	r := servicecomb.NewSCRegistry(scClient)

	wg.Add(2)
	go func() {
		defer wg.Done()
		addr := "127.0.0.1:8701"
		h := server.Default(
			server.WithHostPorts(addr),
			server.WithRegistry(r, &registry.Info{
				ServiceName: "custom-config-demo",
				Addr:        utils.NewNetAddr("tcp", addr),
				Weight:      10,
				Tags: map[string]string{
					"key1": "val1",
				},
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
		addr := "127.0.0.1:8702"
		h := server.Default(
			server.WithHostPorts(addr),
			server.WithRegistry(r, &registry.Info{
				ServiceName: "custom-config-demo",
				Addr:        utils.NewNetAddr("tcp", addr),
				Weight:      10,
				Tags: map[string]string{
					"key1": "val1",
				},
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

	wg.Wait()
}
