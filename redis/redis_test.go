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

package redis

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

var redisCli *redis.Client
var ctx = context.Background()

func init() {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	redisCli = rdb
}

// TestRegister Test the Registry in registry.go
func TestRegister(t *testing.T) {
	tests := []struct {
		info    []*registry.Info
		wantErr bool
	}{
		{
			// set single info
			info: []*registry.Info{
				{
					ServiceName: "hertz.test.demo1",
					Addr:        utils.NewNetAddr(tcp, "127.0.0.1:8888"),
					Weight:      10,
					Tags:        nil,
				},
			},
			wantErr: false,
		},
		{
			// set multi infos
			info: []*registry.Info{
				{
					ServiceName: "hertz.test.demo2",
					Addr:        utils.NewNetAddr(tcp, "127.0.0.1:9000"),
					Weight:      15,
					Tags:        nil,
				},
				{
					ServiceName: "hertz.test.demo2",
					Addr:        utils.NewNetAddr(tcp, "127.0.0.1:9001"),
					Weight:      20,
					Tags:        nil,
				},
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		r := NewRedisRegistry("127.0.0.1:6379")
		for _, info := range test.info {
			if err := r.Register(info); err != nil {
				t.Errorf("info register err")
			}
			hash, err := prepareRegistryHash(info)
			assert.False(t, err != nil)
			val := redisCli.HGet(ctx, hash.key, hash.field).Val()
			ri := &registryInfo{}
			err = json.Unmarshal([]byte(val), ri)
			assert.False(t, err != nil)
			assert.Equal(t, info.ServiceName, ri.ServiceName)
			assert.Equal(t, info.Addr.String(), ri.Addr)
			assert.Equal(t, info.Weight, ri.Weight)
			assert.Equal(t, info.Tags, ri.Tags)
		}
	}
}

// TestResolve Test the Resolver in resolver.go
func TestResolve(t *testing.T) {
	type args struct {
		Addr   string
		Weight int
		Tags   map[string]string
	}
	type info struct {
		ServiceName string
		Args        []args
	}
	tests := []struct {
		info    *info
		wantErr bool
	}{
		{
			// test one args
			info: &info{
				ServiceName: "demo1.hertz.local",
				Args: []args{
					{
						Addr:   "127.0.0.1:8888",
						Weight: 10,
						Tags:   map[string]string{"hello": "world"},
					},
				},
			},
			wantErr: false,
		},
		{
			// test multi args
			info: &info{
				ServiceName: "demo2.hertz.local",
				Args: []args{
					{
						Addr:   "127.0.0.1:9001",
						Weight: 10,
						Tags:   map[string]string{"cloudwego": "hertz"},
					},
					{
						Addr:   "127.0.0.1:9000",
						Weight: 15,
						Tags:   map[string]string{"foo": "bar"},
					},
				},
			},
			wantErr: false,
		},
		{
			// test none args
			info: &info{
				ServiceName: "demo3.hertz.local",
				Args:        []args{},
			},
			wantErr: false,
		},
	}
	for _, test := range tests {
		for _, arg := range test.info.Args {
			hash, err := prepareRegistryHash(&registry.Info{
				ServiceName: test.info.ServiceName,
				Addr:        utils.NewNetAddr(tcp, arg.Addr),
				Weight:      arg.Weight,
				Tags:        arg.Tags,
			})
			assert.False(t, err != nil)
			redisCli.HSet(ctx, hash.key, hash.field, hash.value)
		}
		r := NewRedisResolver("127.0.0.1:6379")
		res, err := r.Resolve(context.Background(), test.info.ServiceName)
		assert.False(t, err != nil)
		if len(res.Instances) == 0 {
			assert.Equal(t, res.CacheKey, test.info.ServiceName)
			continue
		}
		assert.Equal(t, res.CacheKey, test.info.ServiceName)
		addr := make(map[string]struct{})
		weight := make(map[int]struct{})
		for _, arg := range test.info.Args {
			addr[arg.Addr] = struct{}{}
			weight[arg.Weight] = struct{}{}
		}
		for _, ins := range res.Instances {
			_, addrOK := addr[ins.Address().String()]
			_, weightOK := weight[ins.Weight()]
			assert.True(t, addrOK)
			assert.True(t, weightOK)
		}
	}
}

// TestNewMentor test singleton
func TestNewMentor(t *testing.T) {
	m1 := newMentor()
	m2 := newMentor()
	assert.Equal(t, m1, m2)
}

// TestForm test form operation
func TestForm(t *testing.T) {
	m := newMentor()
	m.insertForm("hertz", "127.0.0.1:8000")
	m.insertForm("hertz", "127.0.0.1:8001")
	m.insertForm("cloudwego", "127.0.0.1:9999")
	assert.Equal(t, map[string]addrs{
		"hertz":     {"127.0.0.1:8000", "127.0.0.1:8001"},
		"cloudwego": {"127.0.0.1:9999"},
	}, m.mform)
	m.removeService("cloudwego")
	m.removeAddr("hertz", "127.0.0.1:8001")
	assert.Equal(t, map[string]addrs{
		"hertz": {"127.0.0.1:8000"},
	}, m.mform)
}
