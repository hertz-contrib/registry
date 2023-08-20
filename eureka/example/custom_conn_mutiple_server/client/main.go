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
	"encoding/json"
	"net"
	"time"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/cloudwego/hertz/pkg/app/client/loadbalance"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/hertz-contrib/registry/eureka"
	"github.com/hudl/fargo"
)

type Message struct {
	Message string `json:"message"`
}

func main() {
	// build a eureka resolver from custom conn
	eurekaConn := fargo.EurekaConnection{
		ServiceUrls:  []string{"http://127.0.0.1:8761/eureka"},
		Timeout:      10 * time.Second,
		Retries:      3,
		PollInterval: 30 * time.Second,
	}
	r := eureka.NewEurekaResolverFromConn(eurekaConn)

	discoveryWithSD(r)
	discoveryWithTag(r)
	discoveryWithCustomizedAddr(r)
	discoveryWithLoadBalanceOptions(r)
	discoveryThenUsePostMethod(r)
}

func discoveryWithSD(r discovery.Resolver) {
	hlog.Info("simply discovery:")
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
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

func discoveryWithTag(r discovery.Resolver) {
	hlog.Info("discovery with tag:")
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	cli.Use(sd.Discovery(r))

	for i := 0; i < 10; i++ {
		// set request method、config、request uri
		req := protocol.AcquireRequest()
		req.SetOptions(config.WithSD(true), config.WithTag("key1", "val1"))
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

func discoveryWithCustomizedAddr(r discovery.Resolver) {
	hlog.Info("discovery with customizedAddr:")
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}

	cli.Use(sd.Discovery(r, sd.WithCustomizedAddrs(net.JoinHostPort("127.0.0.1", "5001"))))
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

func discoveryWithLoadBalanceOptions(r discovery.Resolver) {
	hlog.Info("discovery with loadBalanceOptions:")
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	cli.Use(sd.Discovery(r, sd.WithLoadBalanceOptions(loadbalance.NewWeightedBalancer(), loadbalance.Options{
		RefreshInterval: 5 * time.Second,
		ExpireInterval:  15 * time.Second,
	})))
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

func discoveryThenUsePostMethod(r discovery.Resolver) {
	hlog.Info("discovery and use post method to send request:")
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	cli.Use(sd.Discovery(r))

	for i := 0; i < 10; i++ {
		// set request config、method、request uri.
		req := protocol.AcquireRequest()
		req.SetOptions(config.WithSD(true))
		req.SetMethod("POST")
		req.SetRequestURI("http://hertz.discovery.eureka/ping")
		m := Message{Message: "hello"}
		bytes, _ := json.Marshal(m)
		// set body and content type
		req.SetBody(bytes)
		req.Header.SetContentTypeBytes([]byte("application/json"))
		resp := protocol.AcquireResponse()
		// send request
		err := cli.Do(context.Background(), req, resp)
		if err != nil {
			hlog.Fatal(err)
		}
		hlog.Infof("code=%d,body=%s", resp.StatusCode(), string(resp.Body()))
	}
}
