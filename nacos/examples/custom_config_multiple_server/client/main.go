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
	"encoding/json"
	"time"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/cloudwego/hertz/pkg/app/client/loadbalance"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/hertz-contrib/registry/nacos"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
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
	nacosCli, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		})
	if err != nil {
		panic(err)
	}
	r := nacos.NewNacosResolver(nacosCli)

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
		status, body, err := cli.Get(context.Background(), nil, "http://hertz.custom-config.demo/ping", config.WithSD(true))
		if err != nil {
			hlog.Fatal(err)
		}
		hlog.Infof("code=%d,body=%s", status, string(body))
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
		status, body, err := cli.Get(context.Background(), nil, "http://hertz.custom-config.demo/ping",
			config.WithSD(true),
			config.WithTag("key1", "val1"))
		if err != nil {
			hlog.Fatal(err)
		}
		hlog.Infof("code=%d,body=%s", status, string(body))
	}
}

func discoveryWithCustomizedAddr(r discovery.Resolver) {
	hlog.Info("discovery with customizedAddr:")
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}

	cli.Use(sd.Discovery(r, sd.WithCustomizedAddrs("127.0.0.1:8088")))
	for i := 0; i < 10; i++ {
		status, body, err := cli.Get(context.Background(), nil, "http://hertz.custom-config.demo/ping", config.WithSD(true))
		if err != nil {
			hlog.Fatal(err)
		}
		hlog.Infof("code=%d,body=%s", status, string(body))
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
		status, body, err := cli.Get(context.Background(), nil, "http://hertz.custom-config.demo/ping", config.WithSD(true))
		if err != nil {
			hlog.Fatal(err)
		}
		hlog.Infof("code=%d,body=%s", status, string(body))
	}
}

func discoveryThenUsePostMethod(r discovery.Resolver) {
	hlog.Info("discovery and use post method to send request:")
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	cli.Use(sd.Discovery(r))

	// set request config、method、request uri.
	req := protocol.AcquireRequest()
	req.SetOptions(config.WithSD(true))
	req.SetMethod("POST")
	req.SetRequestURI("http://hertz.custom-config.demo/hello")
	message := Message{Message: "hello", Name: "a handsome man"}
	bytes, _ := json.Marshal(message)
	// set body and content type
	req.SetBody(bytes)
	req.Header.SetContentTypeBytes([]byte("application/json"))
	resp := protocol.AcquireResponse()
	// send request
	err = cli.Do(context.Background(), req, resp)
	if err != nil {
		hlog.Fatal(err)
	}
	hlog.Infof("code=%d,body=%s", resp.StatusCode(), string(resp.Body()))
}
