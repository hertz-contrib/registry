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

package consul

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

var (
	cRegistry   registry.Registry
	cResolver   discovery.Resolver
	consulAddr  = "127.0.0.1:8500"
	localIpAddr string
)

func init() {
	config1 := consulapi.DefaultConfig()
	config1.Address = consulAddr
	cli1, err := consulapi.NewClient(config1)
	if err != nil {
		log.Fatal(err)
		return
	}
	cRegistry = NewConsulRegister(cli1)

	config2 := consulapi.DefaultConfig()
	config2.Address = consulAddr
	cli2, err := consulapi.NewClient(config2)
	if err != nil {
		log.Fatal(err)
		return
	}
	cResolver = NewConsulResolver(cli2)

	localIpAddr, err = getLocalIPv4Address()
	if err != nil {
		log.Fatal(err)
	}
}

func getLocalIPv4Address() (string, error) {
	addr, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addr {
		ipNet, isIpNet := addr.(*net.IPNet)
		if isIpNet && !ipNet.IP.IsLoopback() {
			ipv4 := ipNet.IP.To4()
			if ipv4 != nil {
				return ipv4.String(), nil
			}
		}
	}
	return "", fmt.Errorf("not found ipv4 address")
}

// TestNewConsulResolver tests unit test preparatory work.
func TestConsulPrepared(t *testing.T) {
	assert.NotNil(t, cRegistry)
	assert.NotNil(t, cResolver)
	assert.NotEmpty(t, localIpAddr)
}

// TestNewConsulRegister tests the NewConsulRegister function.
func TestNewConsulRegister(t *testing.T) {
	t.Parallel()
	config := consulapi.DefaultConfig()
	config.Address = consulAddr
	cli, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal(err)
		return
	}
	consulRegister := NewConsulRegister(cli)
	assert.NotNil(t, consulRegister)
}

// TestNewConsulRegisterWithCheckOption tests the NewConsulRegister function with check option.
func TestNewConsulRegisterWithCheckOption(t *testing.T) {
	t.Parallel()
	config := consulapi.DefaultConfig()
	config.Address = consulAddr
	cli, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal(err)
		return
	}

	check := new(consulapi.AgentServiceCheck)
	check.Timeout = "10s"
	check.Interval = "10s"
	check.DeregisterCriticalServiceAfter = "1m"

	consulResolver := NewConsulRegister(cli, WithCheck(check))
	assert.NotNil(t, consulResolver)
}

// TestNewConsulResolver tests the NewConsulResolver function .
func TestNewConsulResolver(t *testing.T) {
	t.Parallel()
	config := consulapi.DefaultConfig()
	config.Address = consulAddr
	cli, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal(err)
		return
	}

	consulResolver := NewConsulResolver(cli)
	assert.NotNil(t, consulResolver)
}

// TestConsulRegister tests the Register function with Hertz.
func TestConsulRegister(t *testing.T) {
	t.Parallel()
	config := consulapi.DefaultConfig()
	config.Address = consulAddr
	consulClient, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal(err)
		return
	}

	var (
		testSvcName   = "hertz.test.demo1"
		testSvcPort   = fmt.Sprintf("%d", 8581)
		testSvcAddr   = net.JoinHostPort(localIpAddr, testSvcPort)
		testSvcWeight = 777
	)

	r := NewConsulRegister(consulClient)
	h := server.Default(
		server.WithHostPorts(testSvcAddr),
		server.WithRegistry(r, &registry.Info{
			ServiceName: testSvcName,
			Addr:        utils.NewNetAddr("tcp", testSvcAddr),
			Weight:      testSvcWeight,
		}),
	)

	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"ping": "pong1"})
	})
	go h.Spin()

	// wait for health check passing
	time.Sleep(time.Second * 6)

	list, _, err := consulClient.Health().Service(testSvcName, "", true, nil)
	assert.Nil(t, err)
	if assert.Equal(t, 1, len(list)) {
		ss := list[0]
		gotSvc := ss.Service
		assert.Equal(t, testSvcName, gotSvc.Service)
		assert.Equal(t, testSvcAddr, net.JoinHostPort(gotSvc.Address, fmt.Sprintf("%d", gotSvc.Port)))
		assert.Equal(t, testSvcWeight, gotSvc.Weights.Passing)
	}
}

// TestConsulDiscovery tests the ConsulDiscovery function with Hertz.
func TestConsulDiscovery(t *testing.T) {
	t.Parallel()
	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = consulAddr
	consulClient, err := consulapi.NewClient(consulConfig)
	if err != nil {
		log.Fatal(err)
		return
	}

	var (
		testSvcName   = "hertz.test.demo2"
		testSvcPort   = fmt.Sprintf("%d", 8582)
		testSvcAddr   = net.JoinHostPort(localIpAddr, testSvcPort)
		testSvcWeight = 777
		metaList      = map[string]string{
			"k1": "vv1",
			"k2": "vv2",
			"k3": "vv3",
		}
	)

	r := NewConsulRegister(consulClient)
	h := server.Default(
		server.WithHostPorts(testSvcAddr),
		server.WithRegistry(r, &registry.Info{
			ServiceName: testSvcName,
			Addr:        utils.NewNetAddr("tcp", testSvcAddr),
			Weight:      testSvcWeight,
			Tags:        metaList,
		}),
	)

	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"ping": "pong1"})
	})
	go h.Spin()

	// wait for health check passing
	time.Sleep(time.Second * 6)

	// build a hertz client with the consul resolver
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	cli.Use(sd.Discovery(cResolver))
	status, body, err := cli.Get(context.Background(), nil, "http://hertz.test.demo2/ping", config.WithSD(true))
	if err != nil {
		hlog.Fatal(err)
	}
	assert.Equal(t, "{\"ping\":\"pong1\"}", string(body))
	assert.Equal(t, 200, status)
}

// TestConsulDeregister tests the Deregister function with Hertz
func TestConsulDeregister(t *testing.T) {
	t.Parallel()
	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = consulAddr
	consulClient, err := consulapi.NewClient(consulConfig)
	if err != nil {
		log.Fatal(err)
		return
	}

	var (
		testSvcName   = "hertz.test.demo3"
		testSvcPort   = fmt.Sprintf("%d", 8583)
		testSvcAddr   = net.JoinHostPort(localIpAddr, testSvcPort)
		testSvcWeight = 777
		metaList      = map[string]string{
			"k1": "vv1",
			"k2": "vv2",
			"k3": "vv3",
		}
		ctx          = context.Background()
		registryInfo = &registry.Info{
			ServiceName: testSvcName,
			Addr:        utils.NewNetAddr("tcp", testSvcAddr),
			Weight:      testSvcWeight,
			Tags:        metaList,
		}
	)

	r := NewConsulRegister(consulClient)
	h := server.Default(
		server.WithHostPorts(testSvcAddr),
		server.WithRegistry(r, registryInfo),
	)

	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"ping": "pong1"})
	})
	go h.Spin()

	// wait for health check passing
	time.Sleep(time.Second * 6)

	// resolve
	result, err := cResolver.Resolve(context.Background(), testSvcName)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result.Instances))

	err = h.Shutdown(ctx)
	if err != nil {
		log.Fatal(err)
		return
	}

	// wait for hertz to deregister
	time.Sleep(time.Second * 2)

	// resolve again
	result, err = cResolver.Resolve(ctx, testSvcName)
	assert.Nil(t, err)
	assert.Len(t, result.Instances, 0)
}
