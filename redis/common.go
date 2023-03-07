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
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/go-redis/redis/v8"
)

const (
	Redis      = "redis"
	register   = "register"
	deregister = "deregister"
	hertz      = "hertz"
	server     = "server"
	tcp        = "tcp"
)

const (
	defaultExpireTime    = time.Second * 60
	defaultTickerTime    = time.Second * 30
	defaultKeepAliveTime = time.Second * 60
	defaultMonitorTime   = time.Second * 30
	defaultWeight        = 10
)

type Option func(opts *redis.Options)

func WithPassword(password string) Option {
	return func(opts *redis.Options) {
		opts.Password = password
	}
}

func WithDB(db int) Option {
	return func(opts *redis.Options) {
		opts.DB = db
	}
}

func WithTLSConfig(t *tls.Config) Option {
	return func(opts *redis.Options) {
		opts.TLSConfig = t
	}
}

func WithDialer(dialer func(ctx context.Context, network, addr string) (net.Conn, error)) Option {
	return func(opts *redis.Options) {
		opts.Dialer = dialer
	}
}

func WithReadTimeout(t time.Duration) Option {
	return func(opts *redis.Options) {
		opts.ReadTimeout = t
	}
}

func WithWriteTimeout(t time.Duration) Option {
	return func(opts *redis.Options) {
		opts.WriteTimeout = t
	}
}

type registryHash struct {
	key   string
	field string
	value string
}

type registryInfo struct {
	ServiceName string            `json:"service_name"`
	Addr        string            `json:"addr"`
	Weight      int               `json:"weight"`
	Tags        map[string]string `json:"tags"`
}

func validateRegistryInfo(info *registry.Info) error {
	if info == nil {
		return fmt.Errorf("registry.Info can not be empty")
	}
	if info.ServiceName == "" {
		return fmt.Errorf("registry.Info ServiceName can not be empty")
	}
	if info.Addr == nil {
		return fmt.Errorf("registry.Info Addr can not be empty")
	}
	return nil
}

func generateKey(serviceName, serviceType string) string {
	return fmt.Sprintf("/%s/%s/%s", hertz, serviceName, serviceType)
}

func generateMsg(msgType, serviceName, serviceAddr string) string {
	return fmt.Sprintf("%s-%s-%s", msgType, serviceName, serviceAddr)
}

func prepareRegistryHash(info *registry.Info) (*registryHash, error) {
	meta, err := sonic.Marshal(convertInfo(info))
	if err != nil {
		return nil, err
	}
	return &registryHash{
		key:   generateKey(info.ServiceName, server),
		field: info.Addr.String(),
		value: string(meta),
	}, nil
}

func convertInfo(info *registry.Info) *registryInfo {
	return &registryInfo{
		ServiceName: info.ServiceName,
		Addr:        info.Addr.String(),
		Weight:      info.Weight,
		Tags:        info.Tags,
	}
}

func keepAlive(ctx context.Context, hash *registryHash, r *redisRegistry) {
	ticker := time.NewTicker(defaultTickerTime)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			r.client.Expire(ctx, hash.key, defaultKeepAliveTime)
		case <-ctx.Done():
			break
		}
	}
}
