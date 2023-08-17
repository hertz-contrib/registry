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

	"github.com/cloudwego/hertz/pkg/protocol"

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
		// set request method、config、request uri
		req := protocol.AcquireRequest()
		req.SetOptions(config.WithSD(true))
		req.SetMethod("GET")
		req.SetRequestURI("http://hertz.discovery.eureka/ping")
		// set content type
		req.Header.SetContentTypeBytes([]byte("application/json"))
		resp := protocol.AcquireResponse()
		// send request
		err = cli.Do(context.Background(), req, resp)
		if err != nil {
			hlog.Fatal(err)
		}
		hlog.Infof("code=%d,body=%s", resp.StatusCode(), string(resp.Body()))
	}
}
