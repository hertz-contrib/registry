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

package servicecomb

import (
	"context"
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

const scAddr = "127.0.0.1:30100"

// TestServiceCombRegistryWithHertz Test servicecomb registry complete with hertz.
func TestServiceCombRegistryWithHertz(t *testing.T) {
	addr := "127.0.0.1:8710"
	serviceName := "hertz.servicecomb.test"
	r, err := NewDefaultSCRegistry([]string{scAddr})
	assert.Nil(t, err)
	h := server.Default(
		server.WithHostPorts(addr),
		server.WithRegistry(r, &registry.Info{
			ServiceName: serviceName,
			Addr:        utils.NewNetAddr("tcp", addr),
			Weight:      10,
			Tags:        nil,
		}),
	)

	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"ping": "pong1"})
	})

	go h.Spin()

	time.Sleep(3 * time.Second)

	scResolver, err := NewDefaultSCResolver([]string{scAddr})
	assert.Nil(t, err)

	cli, err := client.NewClient()
	assert.Nil(t, err)

	cli.Use(sd.Discovery(scResolver))
	accessUrl := "http://" + serviceName + "/ping"
	status, body, err := cli.Get(context.Background(), nil, accessUrl, config.WithSD(true))
	assert.Nil(t, err)
	assert.Equal(t, 200, status)
	assert.Equal(t, "{\"ping\":\"pong1\"}", string(body))

	// compare data
	opt := h.GetOptions()
	assert.Equal(t, opt.RegistryInfo.Weight, 10)
	assert.Equal(t, opt.RegistryInfo.Addr.String(), addr)
	assert.Equal(t, opt.RegistryInfo.ServiceName, serviceName)
	assert.Nil(t, opt.RegistryInfo.Tags)

	_ = h.Shutdown(context.Background())
	time.Sleep(5 * time.Second)

	status1, body1, err1 := cli.Get(context.Background(), nil, accessUrl, config.WithSD(true))
	assert.NotNil(t, err1)
	assert.Equal(t, 0, status1)
	assert.Equal(t, "", string(body1))
}

// TestServiceCombDiscovery Test servicecomb registry and deregistry.
func TestServiceCombDiscovery(t *testing.T) {
	// register
	r, err := NewDefaultSCRegistry([]string{scAddr})
	assert.Nil(t, err)
	serviceName := "discovery.test"
	tags := map[string]string{"group": "blue", "idc": "hd1"}
	addr := utils.NewNetAddr("tcp", ":8711")
	info := &registry.Info{ServiceName: serviceName, Weight: 100, Tags: tags, Addr: addr}
	err = r.Register(info)
	assert.Nil(t, err)
	time.Sleep(3 * time.Second)

	// resolve
	res, err := NewDefaultSCResolver([]string{scAddr})
	assert.Nil(t, err)
	target := res.Target(context.Background(), &discovery.TargetInfo{Host: serviceName, Tags: nil})
	result, err := res.Resolve(context.Background(), target)
	assert.Nil(t, err)

	// compare data
	if len(result.Instances) == 1 {
		instance := result.Instances[0]

		host, port, err := net.SplitHostPort(instance.Address().String())
		assert.Nil(t, err)
		local := utils.LocalIP()

		if host != local {
			t.Errorf("instance host is mismatch, expect: %s, in fact: %s", local, host)
		}
		if port != "8711" {
			t.Errorf("instance port is mismatch, expect: %s, in fact: %s", "9999", port)
		}
		for k, v := range info.Tags {
			if v1, exist := instance.Tag(k); !exist || v != v1 {
				t.Errorf("instance tags is mismatch, expect k:v %s:%s, in fact k:v %s:%s", k, v, k, v1)
			}
		}
	} else {
		t.Errorf("instance num mismatch, expect: %d, in fact: %d", 1, len(result.Instances))
	}

	// deregister
	err = r.Deregister(info)
	assert.Nil(t, err)
	time.Sleep(3 * time.Second)

	// resolve again
	result, err = res.Resolve(context.Background(), target)
	assert.Nil(t, err)
	assert.Empty(t, result.Instances)
	assert.Equal(t, serviceName, result.CacheKey)
}

func TestParseAddr(t *testing.T) {

}
