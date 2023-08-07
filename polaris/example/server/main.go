/*
 * Copyright 2021 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

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
	Namespace = "basic"
	// At present,polaris server tag is v1.4.0，can't support auto create namespace,
	// If you want to use a namespace other than basic,Polaris ,before you register an instance,
	// you should create the namespace at polaris console first.
)

func main() {
	r, err := polaris.NewPolarisRegistry()
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
