package etcd

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cloudwego/hertz/pkg/common/utils"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
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
			info: []*registry.Info{
				{
					ServiceName: "hertz.test.demo2",
					Addr:        utils.NewNetAddr("tcp", "127.0.0.1:8000"),
					Weight:      10,
					Tags:        nil,
				},
				{
					ServiceName: "hertz.test.demo2",
					Addr:        utils.NewNetAddr("tcp", "127.0.0.1:8001"),
					Weight:      10,
					Tags:        nil,
				},
			},
			wantErr: false,
		},
	}
	for _, tes := range tests {
		r, err := NewEtcdRegistry(etcdCli, timeout)
		assert.False(t,err != nil)
		for _, info := range tes.info {
			r.Register(info)
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			path, err := buildPath(info)

			assert.False(t, err != nil)

			kv, err := etcdCli.Get(ctx, path)
			cancel()

			assert.False(t, err != nil)
			assert.False(t, len(kv.Kvs) == 0 || len(kv.Kvs) > 1)

			val := kv.Kvs[0].Value
			en := new(RegistryEntity)
			json.Unmarshal(val, en)

			assert.Equal(t, en.Tags, info.Tags)
			assert.Equal(t, en.Weight, info.Weight)
		}
	}
}
