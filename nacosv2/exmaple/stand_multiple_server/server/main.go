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
	"log"
	"sync"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/registry/nacosv2"
)

var (
	wg        sync.WaitGroup
	server1IP = "127.0.0.1:8088"
	server2IP = "127.0.0.1:8089"
)

func main() {
	r, err := nacosv2.NewDefaultNacosV2Registry()
	if err != nil {
		log.Fatal(err)
		return
	}
	wg.Add(2)
	go func() {
		defer wg.Done()
		h := server.Default(
			server.WithHostPorts(server1IP),
			server.WithRegistry(r, &registry.Info{
				ServiceName: "hertz.test.demo",
				Addr:        utils.NewNetAddr("tcp", server1IP),
				Weight:      10,
				Tags:        nil,
			}),
		)
		h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
			ctx.JSON(consts.StatusOK, utils.H{"ping1": "pong1"})
		})
		h.Spin()
	}()

	go func() {
		defer wg.Done()
		h := server.Default(
			server.WithHostPorts(server2IP),
			server.WithRegistry(r, &registry.Info{
				ServiceName: "hertz.test.demo",
				Addr:        utils.NewNetAddr("tcp", server2IP),
				Weight:      10,
				Tags:        nil,
			}),
		)
		h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
			ctx.JSON(consts.StatusOK, utils.H{"ping2": "pong2"})
		})
		h.Spin()
	}()

	wg.Wait()
}
