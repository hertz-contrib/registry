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
	"net"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/stretchr/testify/assert"
)

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
}
