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

package nacos

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/cloudwego/hertz/pkg/app/client/loadbalance"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/stretchr/testify/assert"
)

var namingClient = getNamingClient()

// getNamingClient use to config for naming_client by default.
func getNamingClient() naming_client.INamingClient {
	// create ServerConfig
	sc := []constant.ServerConfig{
		*constant.NewServerConfig("127.0.0.1", 8848, constant.WithContextPath("/nacos")),
	}

	// create ClientConfig
	cc := *constant.NewClientConfig(
		constant.WithCustomLogger(nil),
		constant.WithNamespaceId(""),
		constant.WithTimeoutMs(50000),
		constant.WithUpdateCacheWhenEmpty(true),
		constant.WithNotLoadCacheAtStart(true),
	)

	// create naming client
	newClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		panic(err)
	}
	return newClient
}

// TestRegistryAndDeregister use to test Register, Deregister, NewNacosRegistry.
func TestRegistryAndDeregister(t *testing.T) {
	register := NewNacosRegistry(namingClient)
	infos := []registry.Info{
		{
			ServiceName: "service-name",
			Addr:        utils.NewNetAddr("tcp", "10.4.0.1:8849"),
			Weight:      10,
			Tags:        nil,
		},
		{
			ServiceName: "",
			Addr:        utils.NewNetAddr("tcp", "10.4.0.1:8849"),
			Weight:      -4654564512,
			Tags:        nil,
		},
		{
			ServiceName: "service-name",
			Addr:        utils.NewNetAddr("tcp", "&6^#$%^#"),
			Weight:      10,
			Tags:        nil,
		},
		{
			ServiceName: "service-name",
			Addr:        utils.NewNetAddr("tcp", "10.4.0.45.1:&(#"),
			Weight:      10,
			Tags:        nil,
		},
		{
			ServiceName: "service-name",
			Addr:        nil,
			Weight:      10,
			Tags:        nil,
		},
		{
			ServiceName: "service-name",
			Addr:        utils.NewNetAddr("tcp", "10.4.0.45.1:884959829651"),
			Weight:      10,
			Tags:        nil,
		},
	}
	err := register.Register(&infos[0])
	assert.Nil(t, err)
	time.Sleep(time.Second)
	err = register.Deregister(&infos[0])
	assert.Nil(t, err)

	// registry info error
	err = register.Register(&infos[1])
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "valid parse registry info error")

	// registry info addr error
	err = register.Register(&infos[2])
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "parse registry info addr error")

	// port error
	err = register.Register(&infos[3])
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "parse registry info port error")

	// addr nil
	err = register.Register(&infos[4])
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "valid parse registry info error")

	// instance error
	err = register.Register(&infos[5])
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "register instance error")
}

// TestNewDefaultResolver test new a default nacos resolver.
func TestNewDefaultResolver(t *testing.T) {
	r, err := NewDefaultNacosResolver()
	assert.NotNil(t, r)
	assert.Nil(t, err)
}

// TestMultiInstancesWithDefRegistry  use DefaultNacosRegistry to test registry multiple service,then deregister one
func TestMultiInstancesWithDefRegistry(t *testing.T) {
	var (
		svcName     = "MultipleInstances"
		clusterName = "TheCluster"
		groupName   = "TheGroup"
	)
	got, err := NewDefaultNacosRegistry(WithRegistryCluster(clusterName), WithRegistryGroup(groupName))
	assert.Nil(t, err)

	time.Sleep(time.Second)
	err = got.Register(&registry.Info{
		ServiceName: svcName,
		Addr:        utils.NewNetAddr("tcp", "127.0.0.1:8081"),
	})
	assert.Nil(t, err, "first register error")

	err = got.Register(&registry.Info{
		ServiceName: svcName,
		Addr:        utils.NewNetAddr("tcp", "127.0.0.1:8082"),
	})
	assert.Nil(t, err, "second register error")

	err = got.Register(&registry.Info{
		ServiceName: svcName,
		Addr:        utils.NewNetAddr("tcp", "127.0.0.1:8083"),
	})
	assert.Nil(t, err, "third register error")

	time.Sleep(time.Second * 1)
	newclient := getNamingClient()
	res, err := newclient.SelectAllInstances(vo.SelectAllInstancesParam{
		ServiceName: svcName,
		GroupName:   groupName,
		Clusters:    []string{clusterName},
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(res), "successful register not three")

	time.Sleep(time.Second)
	err = got.Deregister(&registry.Info{
		ServiceName: svcName,
		Addr:        utils.NewNetAddr("tcp", "127.0.0.1:8083"),
	})
	assert.Nil(t, err)
	time.Sleep(time.Second * 3)
	res, err = newclient.SelectAllInstances(vo.SelectAllInstancesParam{
		ServiceName: svcName,
		GroupName:   groupName,
		Clusters:    []string{clusterName},
	})

	assert.Nil(t, err)
	if assert.Equal(t, 2, len(res), "deregister one, instances num should be two") {
		for _, i := range res {
			assert.Equal(t, "127.0.0.1", i.Ip)
			assert.Contains(t, []uint64{8081, 8082}, i.Port)
		}
	}
}

// TestMultipleInstances test registry multiple service,then deregister one
func TestMultipleInstances(t *testing.T) {
	var (
		svcName     = "MultipleInstances"
		clusterName = "TheCluster"
		groupName   = "TheGroup"
	)

	time.Sleep(time.Second)
	got := NewNacosRegistry(namingClient, WithRegistryCluster(clusterName), WithRegistryGroup(groupName))
	if !assert.NotNil(t, got) {
		t.Errorf("err: new registry fail")
		return
	}
	time.Sleep(time.Second)
	err := got.Register(&registry.Info{
		ServiceName: svcName,
		Addr:        utils.NewNetAddr("tcp", "127.0.0.1:8081"),
	})
	assert.Nil(t, err)

	err = got.Register(&registry.Info{
		ServiceName: svcName,
		Addr:        utils.NewNetAddr("tcp", "127.0.0.1:8082"),
	})
	assert.Nil(t, err)

	err = got.Register(&registry.Info{
		ServiceName: svcName,
		Addr:        utils.NewNetAddr("tcp", "127.0.0.1:8083"),
	})
	assert.Nil(t, err)

	time.Sleep(time.Second)
	res, err := namingClient.SelectAllInstances(vo.SelectAllInstancesParam{
		ServiceName: svcName,
		GroupName:   groupName,
		Clusters:    []string{clusterName},
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(res), "successful register not three")

	time.Sleep(time.Second)
	err = got.Deregister(&registry.Info{
		ServiceName: svcName,
		Addr:        utils.NewNetAddr("tcp", "127.0.0.1:8083"),
	})
	assert.Nil(t, err)

	time.Sleep(time.Second * 3)
	res, err = namingClient.SelectAllInstances(vo.SelectAllInstancesParam{
		ServiceName: svcName,
		GroupName:   groupName,
		Clusters:    []string{clusterName},
	})
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(res), "deregister one, instances num should be two") {
		for _, i := range res {
			assert.Equal(t, "127.0.0.1", i.Ip)
			assert.Contains(t, []uint64{8081, 8082}, i.Port)
		}
	}
}

// TestResolverResolve test Resolve a service.
func TestResolverResolve(t *testing.T) {
	h := server.Default(
		server.WithHostPorts("127.0.0.1:8080"),
		server.WithRegistry(NewNacosRegistry(namingClient), &registry.Info{
			ServiceName: "demo.hertz-contrib.local",
			Addr:        utils.NewNetAddr("tcp", "127.0.0.1:8080"),
			Weight:      10,
		}),
	)
	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.String(200, "pong")
	})
	go h.Spin()
	time.Sleep(2 * time.Second)

	type fields struct {
		cli naming_client.INamingClient
	}
	type args struct {
		ctx  context.Context
		desc string
	}
	// some cases
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    discovery.Result
		wantErr bool
	}{
		{
			name: "common",
			args: args{
				ctx:  context.Background(),
				desc: "demo.hertz-contrib.local",
			},
			fields: fields{cli: namingClient},
		},
		{
			name: "wrong desc",
			args: args{
				ctx:  context.Background(),
				desc: "xxxx.yyyy.zzzz",
			},
			fields:  fields{cli: namingClient},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewNacosResolver(tt.fields.cli)
			_, err := n.Resolve(tt.args.ctx, tt.args.desc)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && !strings.Contains(err.Error(), "instance list is empty") {
				t.Errorf("Resolve err is not expected")
				return
			}
		})
	}

	err := NewNacosRegistry(namingClient).Deregister(&registry.Info{
		ServiceName: "demo.hertz-contrib.local",
		Addr:        utils.NewNetAddr("tcp", "127.0.0.1:8080"),
		Weight:      10,
	})
	assert.Nil(t, err)
}

// TestDefaultNacosRegistry use to test NewDefaultNacosRegistry, Register and Deregister.
func TestDefaultNacosRegistry(t *testing.T) {
	register, err := NewDefaultNacosRegistry()
	assert.Nil(t, err)
	info := &registry.Info{
		ServiceName: "service-name",
		Addr:        utils.NewNetAddr("tcp", "10.4.0.1:8849"),
		Weight:      10,
		Tags:        nil,
	}
	err = register.Register(info)
	assert.Nil(t, err)
	err = register.Deregister(info)
	assert.Nil(t, err)
}

// TestHertzAppWithNacosRegistry test a client call a hertz app with NacosRegistry
func TestHertzAppWithNacosRegistry(t *testing.T) {
	register := NewNacosRegistry(namingClient)
	address := "127.0.0.1:4576"
	srvName := "demo.hertz-contrib.testing"
	var opts []config.Option
	opts = append(opts, server.WithHostPorts(address))
	opts = append(opts, server.WithRegistry(register, &registry.Info{
		ServiceName: srvName,
		Addr:        utils.NewNetAddr("tcp", address),
		Weight:      10,
		Tags:        nil,
	}))
	// run a hertz app,registry src info into NacosRegistry
	srv := server.New(opts...)
	srv.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.String(200, "pong")
	})
	go srv.Spin()
	// Because delayed registration, we need sleep more time.
	time.Sleep(2 * time.Second)

	// client call an url, with NacosResolver
	newClient, _ := client.NewClient()
	resolver := NewNacosResolver(namingClient)
	newClient.Use(sd.Discovery(resolver, sd.WithLoadBalanceOptions(
		loadbalance.NewWeightedBalancer(),
		loadbalance.Options{
			ExpireInterval:  3 * time.Second,
			RefreshInterval: 1 * time.Second,
		}),
	))

	status, body, err := newClient.Get(context.TODO(), nil, "http://demo.hertz-contrib.testing/ping",
		config.WithSD(true))
	assert.Nil(t, err)
	assert.Equal(t, 200, status)
	assert.Equal(t, "pong", string(body))

	if err = srv.Shutdown(context.TODO()); err != nil {
		t.Error(err)
	}
	time.Sleep(6 * time.Second)
	status1, body1, err1 := newClient.Get(context.Background(), nil, "http://demo.hertz-contrib.testing/ping",
		config.WithSD(true))
	assert.NotNil(t, err1)
	assert.Equal(t, 0, status1)
	assert.Equal(t, "", string(body1))
}

// TestResolverDifferentGroup test NewResolver WithCluster option
func TestResolverDifferentGroup(t *testing.T) {
	var opts1 []config.Option
	var opts2 []config.Option

	opts1 = append(opts1, server.WithRegistry(NewNacosRegistry(namingClient), &registry.Info{
		ServiceName: "demo.hertz-contrib.test1",
		Addr:        utils.NewNetAddr("tcp", "127.0.0.1:7000"),
		Weight:      10,
		Tags:        nil,
	}))
	opts1 = append(opts1, server.WithHostPorts("127.0.0.1:7000"))
	srv1 := server.New(opts1...)
	srv1.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.String(200, "pong1")
	})

	opts2 = append(opts2, server.WithRegistry(NewNacosRegistry(namingClient, WithRegistryGroup("OTHER")), &registry.Info{
		ServiceName: "demo.hertz-contrib.test1",
		Addr:        utils.NewNetAddr("tcp", "127.0.0.1:7001"),
		Weight:      10,
		Tags:        nil,
	}))
	opts2 = append(opts2, server.WithHostPorts("127.0.0.1:7001"))
	srv2 := server.New(opts2...)
	srv2.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.String(200, "pong2")
	})

	go srv1.Spin()
	go srv2.Spin()

	time.Sleep(2 * time.Second)

	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	r := NewNacosResolver(namingClient)
	cli.Use(sd.Discovery(r))

	ctx, cancelFunc := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelFunc()

	status, body, err := cli.Get(ctx, nil,
		"http://demo.hertz-contrib.test1/ping", config.WithSD(true))
	assert.Nil(t, err)
	assert.Equal(t, 200, status)
	assert.Equal(t, "pong1", string(body))

	cli2, err2 := client.NewClient()
	if err2 != nil {
		panic(err2)
	}

	ctx2, cancelFunc2 := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelFunc2()

	cli2.Use(sd.Discovery(NewNacosResolver(namingClient, WithResolverGroup("OTHER"))))
	status2, body2, err2 := cli2.Get(ctx2, nil,
		"http://demo.hertz-contrib.test1/ping", config.WithSD(true))
	assert.Nil(t, err2)
	assert.Equal(t, 200, status2)
	assert.Equal(t, "pong2", string(body2))
}
