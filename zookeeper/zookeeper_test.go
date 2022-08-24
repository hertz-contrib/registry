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

package zookeeper

import (
	"context"
	"fmt"
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
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/stretchr/testify/assert"
)

func TestZookeeperRegistryAndDeregister(t *testing.T) {
	address := "127.0.0.1:8888"
	r, _ := NewZookeeperRegistry([]string{"127.0.0.1:2181"}, 40*time.Second)
	srvName := "hertz.test.demo"
	h := server.Default(
		server.WithHostPorts(address),
		server.WithRegistry(r, &registry.Info{
			ServiceName: srvName,
			Addr:        utils.NewNetAddr("tcp", address),
			Weight:      10,
			Tags:        nil,
		}))
	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"ping": "pong2"})
	})
	go h.Spin()

	time.Sleep(1 * time.Second)

	// register
	newClient, _ := client.NewClient()
	resolver, _ := NewZookeeperResolver([]string{"127.0.0.1:2181"}, 40*time.Second)
	newClient.Use(sd.Discovery(resolver))

	addr := fmt.Sprintf("http://" + srvName + "/ping")
	status, body, err := newClient.Get(context.Background(), nil, addr, config.WithSD(true))
	assert.Nil(t, err)
	assert.Equal(t, 200, status)
	assert.Equal(t, "{\"ping\":\"pong2\"}", string(body))

	// compare data
	opt := h.GetOptions()
	assert.Equal(t, opt.RegistryInfo.Weight, 10)
	assert.Equal(t, opt.RegistryInfo.Addr.String(), "127.0.0.1:8888")
	assert.Equal(t, opt.RegistryInfo.ServiceName, "hertz.test.demo")
	assert.Nil(t, opt.RegistryInfo.Tags)

	h.Shutdown(context.Background())
	time.Sleep(5 * time.Second)

	status1, body1, err1 := newClient.Get(context.Background(), nil, addr, config.WithSD(true))
	assert.NotNil(t, err1)
	assert.Equal(t, 0, status1)
	assert.Equal(t, "", string(body1))
}

func TestZookeeperDiscovery(t *testing.T) {
	// register
	r, err := NewZookeeperRegistry([]string{"127.0.0.1:2181"}, 40*time.Second)
	assert.Nil(t, err)
	tags := map[string]string{"group": "blue", "idc": "hd1"}
	addr, _ := net.ResolveTCPAddr("tcp", ":9999")
	info := &registry.Info{ServiceName: "product", Weight: 100, Tags: tags, Addr: addr}
	err = r.Register(info)
	assert.Nil(t, err)

	// resolve
	res, err := NewZookeeperResolver([]string{"127.0.0.1:2181"}, 40*time.Second)
	assert.Nil(t, err)
	target := res.Target(context.Background(), &discovery.TargetInfo{Host: "product", Tags: nil})
	result, err := res.Resolve(context.Background(), target)
	assert.Nil(t, err)

	// compare data
	if len(result.Instances) == 0 {
		t.Errorf("instance num mismatch, expect: %d, in fact: %d", 1, 0)
	} else if len(result.Instances) == 1 {
		instance := result.Instances[0]
		host, port, err := net.SplitHostPort(instance.Address().String())
		assert.Nil(t, err)
		local := utils.LocalIP()
		if host != local {
			t.Errorf("instance host is mismatch, expect: %s, in fact: %s", local, host)
		}
		if port != "9999" {
			t.Errorf("instance port is mismatch, expect: %s, in fact: %s", "9999", port)
		}
		if info.Weight != instance.Weight() {
			t.Errorf("instance weight is mismatch, expect: %d, in fact: %d", info.Weight, instance.Weight())
		}
		for k, v := range info.Tags {
			if v1, exist := instance.Tag(k); !exist || v != v1 {
				t.Errorf("instance tags is mismatch, expect k:v %s:%s, in fact k:v %s:%s", k, v, k, v1)
			}
		}
	}

	// deregister
	err = r.Deregister(info)
	assert.Nil(t, err)

	// resolve again
	result, err = res.Resolve(context.Background(), target)
	assert.Nil(t, err)
	assert.Empty(t, result.Instances)
	assert.Equal(t, "product", result.CacheKey)
}

func TestZookeeperResolverWithAuth(t *testing.T) {
	// register
	r, err := NewZookeeperRegistryWithAuth([]string{"127.0.0.1:2181"}, 40*time.Second, "horizon", "horizon")
	assert.Nil(t, err)
	tags := map[string]string{"group": "blue", "idc": "hd1"}
	addr, _ := net.ResolveTCPAddr("tcp", ":9999")
	info := &registry.Info{ServiceName: "product", Weight: 100, Tags: tags, Addr: addr}
	err = r.Register(info)
	assert.Nil(t, err)

	// resolve
	res, err := NewZookeeperResolverWithAuth([]string{"127.0.0.1:2181"}, 40*time.Second, "horizon", "horizon")
	assert.Nil(t, err)
	target := res.Target(context.Background(), &discovery.TargetInfo{Host: "product", Tags: nil})
	result, err := res.Resolve(context.Background(), target)
	assert.Nil(t, err)

	// compare data
	if len(result.Instances) == 0 {
		t.Errorf("instance num mismatch, expect: %d, in fact: %d", 1, 0)
	} else if len(result.Instances) == 1 {
		instance := result.Instances[0]
		host, port, err := net.SplitHostPort(instance.Address().String())
		assert.Nil(t, err)
		local := utils.LocalIP()
		if host != local {
			t.Errorf("instance host is mismatch, expect: %s, in fact: %s", local, host)
		}
		if port != "9999" {
			t.Errorf("instance port is mismatch, expect: %s, in fact: %s", "9999", port)
		}
		if info.Weight != instance.Weight() {
			t.Errorf("instance weight is mismatch, expect: %d, in fact: %d", info.Weight, instance.Weight())
		}
		for k, v := range info.Tags {
			if v1, exist := instance.Tag(k); !exist || v != v1 {
				t.Errorf("instance tags is mismatch, expect k:v %s:%s, in fact k:v %s:%s", k, v, k, v1)
			}
		}
	}

	// deregister
	err = r.Deregister(info)
	assert.Nil(t, err)

	// resolve again
	result, err = res.Resolve(context.Background(), target)
	assert.Nil(t, err)
	assert.Empty(t, result.Instances)
	assert.Equal(t, "product", result.CacheKey)
}
