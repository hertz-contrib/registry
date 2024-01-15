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

package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	hzcli "github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	hzsrv "github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

var (
	redisCli *redis.Client
	ctx      = context.Background()
)

func init() {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	redisCli = rdb
}

// TestRegister Test the Registry in registry.go
func TestRegister(t *testing.T) {
	defer redisCli.FlushDB(ctx)
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
	defer redisCli.FlushDB(ctx)
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

// TestRedisRegistryWithHertz Test redis registry complete workflow (service registry|service de-registry|service resolver) with hertz.
func TestRedisRegistryWithHertz(t *testing.T) {
	addr := "127.0.0.1:8080"
	redisAddr := "127.0.0.1:6379"
	srvName := "hertz.with.registry"
	r := NewRedisRegistry(redisAddr)
	h := hzsrv.Default(
		hzsrv.WithHostPorts(addr),
		hzsrv.WithRegistry(r, &registry.Info{
			ServiceName: srvName,
			Addr:        utils.NewNetAddr("tcp", addr),
			Weight:      10,
			Tags:        nil,
		}),
	)
	h.GET("/ping", func(_ context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"ping": "pong"})
	})
	go h.Spin()
	time.Sleep(4 * time.Second)

	hc, _ := hzcli.NewClient()
	resolver := NewRedisResolver(redisAddr)
	hc.Use(sd.Discovery(resolver))

	url := fmt.Sprintf("http://%v/ping", srvName)
	status, body, err := hc.Get(context.Background(), nil, url, config.WithSD(true))
	assert.Nil(t, err)
	assert.Equal(t, 200, status)
	assert.Equal(t, "{\"ping\":\"pong\"}", string(body))

	opts := h.GetOptions()
	assert.Equal(t, opts.RegistryInfo.ServiceName, srvName)
	assert.Equal(t, opts.RegistryInfo.Addr.String(), addr)
	assert.Equal(t, opts.RegistryInfo.Weight, 10)
	assert.Nil(t, opts.RegistryInfo.Tags)

	if err := h.Shutdown(context.Background()); err != nil {
		t.Errorf("HERTZ: Shutdown error=%v", err)
	}
	time.Sleep(5 * time.Second)

	status2, body2, err2 := hc.Get(context.Background(), nil, url, config.WithSD(true))
	assert.True(t, err2 != nil)
	assert.Equal(t, 0, status2)
	assert.Equal(t, "", string(body2))
}
