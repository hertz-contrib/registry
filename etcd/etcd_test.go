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

package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
)

var (
	etcdCli *clientv3.Client
	timeout time.Duration = time.Second * 2
)

func init() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 2 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	etcdCli = cli
}

// TestRegistry Test the Registry in registry.go
func TestRegistry(t *testing.T) {
	tests := []struct {
		info    []*registry.Info
		wantErr bool
	}{
		{
			// set single info
			info: []*registry.Info{
				{
					ServiceName: "hertz.test.demo1",
					Addr:        utils.NewNetAddr("tcp", "127.0.0.1:8000"),
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
					Addr:        utils.NewNetAddr("tcp", "127.0.0.1:8000"),
					Weight:      15,
					Tags:        nil,
				},
				{
					ServiceName: "hertz.test.demo2",
					Addr:        utils.NewNetAddr("tcp", "127.0.0.1:8001"),
					Weight:      20,
					Tags:        nil,
				},
			},
			wantErr: false,
		},
	}
	for _, tes := range tests {
		r, err := NewEtcdRegistry([]string{"127.0.0.1:2379"})
		assert.False(t, err != nil)
		for _, info := range tes.info {
			if err := r.Register(info); err != nil {
				t.Errorf("info register err")
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)

			assert.False(t, err != nil)
			key := serviceKey(info.ServiceName, info.Addr.String())
			kv, err := etcdCli.Get(ctx, key)

			assert.False(t, err != nil)
			assert.False(t, len(kv.Kvs) == 0 || len(kv.Kvs) > 1)
			cancel()

			val := kv.Kvs[0].Value
			en := new(instanceInfo)
			if err := json.Unmarshal(val, en); err != nil {
				t.Errorf("json unmarshal error")
			}

			assert.Equal(t, en.Tags, info.Tags)
			assert.Equal(t, en.Weight, info.Weight)
		}
	}
}

// TestResolver Test the Resolver in resolver.go
func TestResolver(t *testing.T) {
	type args struct {
		Addr   string
		Weight int
		Tags   map[string]string
	}
	type info struct {
		ServiceName string
		args        []args
	}
	tests := []struct {
		info    *info
		wantErr bool
	}{
		{
			// test one args
			info: &info{
				ServiceName: "demo1.hertz.local",
				args: []args{
					{
						Addr:   "127.0.0.1:8000",
						Weight: 10,
						Tags:   map[string]string{"test": "test1"},
					},
				},
			},
			wantErr: false,
		},
		{
			// test multi args
			info: &info{
				ServiceName: "demo2.hertz.local",
				args: []args{
					{
						Addr:   "127.0.0.1:8000",
						Weight: 2,
						Tags:   map[string]string{"test": "test1"},
					},
					{
						Addr:   "127.0.0.1:8001",
						Weight: 3,
						Tags:   map[string]string{"test": "test2"},
					},
				},
			},
			wantErr: false,
		},
		{
			// test none args
			info: &info{
				ServiceName: "demo3.hertz.local",
				args:        []args{},
			},
			wantErr: false,
		},
	}
	for _, tes := range tests {
		info := tes.info
		// put the addr into the etcd cluster
		for _, args := range tes.info.args {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			key := serviceKey(info.ServiceName, args.Addr)
			addr := utils.NewNetAddr("tcp", args.Addr)
			content, err := json.Marshal(&instanceInfo{
				Network: addr.Network(),
				Address: args.Addr,
				Weight:  args.Weight,
				Tags:    args.Tags,
			})
			if err != nil {
				t.Error(err)
			}
			_, err = etcdCli.Put(ctx, key, string(content))
			if err != nil {
				t.Errorf("path put error")
			}
			cancel()
		}
		r, err := NewEtcdResolver([]string{"127.0.0.1:2379"})
		if err != nil {
			t.Error(err)
		}
		res, err := r.Resolve(context.Background(), tes.info.ServiceName)
		if err != nil {
			t.Error("err found here")
		}
		if len(res.Instances) == 0 {
			assert.Equal(t, res.CacheKey, tes.info.ServiceName)
			continue
		}

		assert.Equal(t, res.CacheKey, tes.info.ServiceName)

		for i, ins := range res.Instances {
			args := tes.info.args[i]

			assert.Equal(t, args.Addr, ins.Address().String())
			assert.Equal(t, args.Weight, ins.Weight())
		}
	}
}

// TestEtcdRegistryWithHertz Test etcd registry complete workflow(service registry|service de-registry|service resolver)with hertz.
func TestEtcdRegistryWithHertz(t *testing.T) {
	address := "127.0.0.1:1234"
	r, _ := NewEtcdRegistry([]string{"127.0.0.1:2379"})
	srvName := "hertz.with.registry"
	h := server.Default(
		server.WithHostPorts(address),
		server.WithRegistry(r, &registry.Info{
			ServiceName: srvName,
			Addr:        utils.NewNetAddr("tcp", address),
			Weight:      10,
			Tags:        nil,
		}))
	h.GET("/ping", func(_ context.Context, ctx *app.RequestContext) {
		ctx.JSON(consts.StatusOK, utils.H{"ping": "pong2"})
	})
	go h.Spin()

	time.Sleep(4 * time.Second)

	// register
	newClient, _ := client.NewClient()
	resolver, _ := NewEtcdResolver([]string{"127.0.0.1:2379"})
	newClient.Use(sd.Discovery(resolver))

	addr := fmt.Sprintf("http://" + srvName + "/ping")
	status, body, err := newClient.Get(context.Background(), nil, addr, config.WithSD(true))
	assert.Nil(t, err)
	assert.Equal(t, 200, status)
	assert.Equal(t, "{\"ping\":\"pong2\"}", string(body))

	// compare data
	opt := h.GetOptions()
	assert.Equal(t, opt.RegistryInfo.Weight, 10)
	assert.Equal(t, opt.RegistryInfo.Addr.String(), "127.0.0.1:1234")
	assert.Equal(t, opt.RegistryInfo.ServiceName, srvName)
	assert.Nil(t, opt.RegistryInfo.Tags)

	if err := h.Shutdown(context.Background()); err != nil {
		t.Errorf("HERTZ: Shutdown error=%v", err)
	}
	time.Sleep(5 * time.Second)

	status1, body1, err1 := newClient.Get(context.Background(), nil, addr, config.WithSD(true))
	assert.True(t, err1 != nil)
	assert.Equal(t, 0, status1)
	assert.Equal(t, "", string(body1))
}

func TestEtcdRegistryWithAddressBlank(t *testing.T) {
	s, endpoint := setupEmbedEtcd(t)
	rg, err := NewEtcdRegistry([]string{endpoint})
	require.Nil(t, err)
	rs, err := NewEtcdResolver([]string{endpoint})
	require.Nil(t, err)

	infoList := []registry.Info{
		{
			ServiceName: "registry-etcd-test",
			Addr:        utils.NewNetAddr("tcp", "[::]:8888"),
			Weight:      27,
			Tags:        map[string]string{"hello": "world"},
		},
		{
			ServiceName: "registry-etcd-test-suffix",
			Addr:        utils.NewNetAddr("tcp", "127.0.0.1:9999"),
			Weight:      27,
			Tags:        map[string]string{"hello": "world"},
		},
	}

	// test register service
	{
		for _, info := range infoList {
			err = rg.Register(&info)
			require.Nil(t, err)

			desc := rs.Target(context.TODO(), &discovery.TargetInfo{
				Host: info.ServiceName,
				Tags: info.Tags,
			})
			result, err := rs.Resolve(context.TODO(), desc)
			require.Nil(t, err)
			address, err := rg.(*etcdRegistry).getAddressOfRegistration(&info)
			require.Nil(t, err)
			expected := discovery.Result{
				CacheKey: info.ServiceName,
				Instances: []discovery.Instance{
					discovery.NewInstance(info.Addr.Network(), address, info.Weight, info.Tags),
				},
			}
			require.Equal(t, expected, result)
		}
	}

	// test deregister service
	{
		for _, info := range infoList {
			err = rg.Deregister(&info)
			assert.Nil(t, err)
		}
	}
	teardownEmbedEtcd(s)
}

func TestEtcdRegistryWithEnvironmentVariable(t *testing.T) {
	s, endpoint := setupEmbedEtcd(t)
	err := os.Setenv(hertzPortToRegistry, "8899")
	if err != nil {
		return
	}
	err = os.Setenv(hertzIpToRegistry, "127.0.0.2")
	if err != nil {
		return
	}

	rg, err := NewEtcdRegistry([]string{endpoint})
	require.Nil(t, err)
	rs, err := NewEtcdResolver([]string{endpoint})
	require.Nil(t, err)

	infoList := []registry.Info{
		{
			ServiceName: "registry-etcd-test",
			Addr:        utils.NewNetAddr("tcp", "[::]:8888"),
			Weight:      27,
			Tags:        map[string]string{"hello": "world"},
		},
		{
			ServiceName: "registry-etcd-test-suffix",
			Addr:        utils.NewNetAddr("tcp", "10.122.1.108:9999"),
			Weight:      27,
			Tags:        map[string]string{"hello": "world"},
		},
	}

	// test register service
	{
		for _, info := range infoList {
			err = rg.Register(&info)
			require.Nil(t, err)

			desc := rs.Target(context.TODO(), &discovery.TargetInfo{
				Host: info.ServiceName,
			})
			result, err := rs.Resolve(context.TODO(), desc)
			require.Nil(t, err)
			address, err := rg.(*etcdRegistry).getAddressOfRegistration(&info)
			require.Nil(t, err)
			expected := discovery.Result{
				CacheKey: info.ServiceName,
				Instances: []discovery.Instance{
					discovery.NewInstance(info.Addr.Network(), address, info.Weight, info.Tags),
				},
			}
			require.Equal(t, expected, result)
		}
	}

	// test deregister service
	{
		for _, info := range infoList {
			err = rg.Deregister(&info)
			require.Nil(t, err)
		}
	}
	os.Unsetenv(hertzPortToRegistry)
	os.Unsetenv(hertzIpToRegistry)
	teardownEmbedEtcd(s)
}

func TestRetryOption(t *testing.T) {
	o := newOptionForServer([]string{"127.0.0.1:2345"})
	assert.Equal(t, o.etcdCfg.Endpoints, []string{"127.0.0.1:2345"})
	assert.Equal(t, uint(5), o.retryCfg.maxAttemptTimes)
	assert.Equal(t, 30*time.Second, o.retryCfg.observeDelay)
	assert.Equal(t, 10*time.Second, o.retryCfg.retryDelay)
}

func TestRetryCustomConfig(t *testing.T) {
	o := newOptionForServer(
		[]string{"127.0.0.1:2345"},
		WithMaxAttemptTimes(10),
		WithObserveDelay(20*time.Second),
		WithRetryDelay(5*time.Second),
	)
	assert.Equal(t, uint(10), o.retryCfg.maxAttemptTimes)
	assert.Equal(t, 20*time.Second, o.retryCfg.observeDelay)
	assert.Equal(t, 5*time.Second, o.retryCfg.retryDelay)
}

func setupEmbedEtcd(t *testing.T) (*embed.Etcd, string) {
	pid := os.Getpid()
	endpoint := fmt.Sprintf("http://localhost:%06d", pid)
	peerEndpoint := fmt.Sprintf("http://localhost:%06d", pid+1)
	u, err := url.Parse(endpoint)
	require.Nil(t, err)
	p, err := url.Parse(peerEndpoint)
	require.Nil(t, err)
	dir, err := os.MkdirTemp("", "etcd_resolver_test")
	require.Nil(t, err)

	cfg := embed.NewConfig()
	cfg.LCUrls = []url.URL{*u}
	cfg.ACUrls = []url.URL{*u}
	cfg.LPUrls = []url.URL{*p}
	cfg.APUrls = []url.URL{*p}
	cfg.Name = "test"
	cfg.InitialCluster = cfg.InitialClusterFromName(cfg.Name)
	// disable etcd log
	cfg.LogLevel = "panic"
	cfg.Dir = dir

	s, err := embed.StartEtcd(cfg)
	require.Nil(t, err)

	<-s.Server.ReadyNotify()
	return s, endpoint
}

func teardownEmbedEtcd(s *embed.Etcd) {
	s.Close()
	_ = os.RemoveAll(s.Config().Dir)
}
