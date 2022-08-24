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
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/clientv3"
)

var (
	etcdCli *clientv3.Client
	timeout time.Duration = 2 * time.Second
	resover discovery.Resolver
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
	r, err := NewEtcdResolver(etcdCli, timeout)
	if err != nil {
		panic(err)
	}
	resover = r
}


func TestResolve(t *testing.T) {
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
		// put the addr into the etcd cluster
		for _, args := range tes.info.args {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			path := Separator + tes.info.ServiceName + Separator + args.Addr
			content, err := json.Marshal(&RegistryEntity{Weight: args.Weight, Tags: args.Tags})
			if err != nil {
				t.Error(err)
			}
			_, err = etcdCli.Put(ctx, path, string(content))
			if err != nil {
				t.Errorf("path put error")
			}
			cancel()
		}
		res, err := resover.Resolve(context.Background(), tes.info.ServiceName)
		if err != nil {

			assert.False(t, tes.wantErr)

			continue
		}

		assert.Equal(t, res.CacheKey, tes.info.ServiceName)

		for i, ins := range res.Instances {
			args := tes.info.args[i]

			assert.Equal(t, args.Addr, ins.Address().String())
			assert.Equal(t, args.Weight, ins.Weight())

			for key, val1 := range args.Tags {
				val2, exist := ins.Tag(key)

				assert.True(t, exist)
				assert.Equal(t, val1, val2)

			}
		}
	}
}
