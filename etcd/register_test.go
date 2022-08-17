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

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/stretchr/testify/assert"
)

func TestBuildPath(t *testing.T) {
	addr := "127.0.0.1:8000"
	path, err := buildPath(&registry.Info{
		ServiceName: "hertz.test.demo",
		Addr:        utils.NewNetAddr("tcp", addr),
		Weight:      10,
		Tags:        nil,
	})
	assert.False(t, err != nil)
	assert.Equal(t, "/hertz.test.demo/127.0.0.1:8000", path)
}

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
		r, err := NewEtcdRegistry(etcdCli, timeout)
		assert.False(t, err != nil)
		for _, info := range tes.info {
			if err := r.Register(info); err != nil {
				t.Errorf("info register err")
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			path, err := buildPath(info)

			assert.False(t, err != nil)

			kv, err := etcdCli.Get(ctx, path)
			cancel()

			assert.False(t, err != nil)
			assert.False(t, len(kv.Kvs) == 0 || len(kv.Kvs) > 1)

			val := kv.Kvs[0].Value
			en := new(RegistryEntity)
			if err := json.Unmarshal(val, en); err != nil {
				t.Errorf("json unmarshal error")
			}

			assert.Equal(t, en.Tags, info.Tags)
			assert.Equal(t, en.Weight, info.Weight)
		}
	}
}
