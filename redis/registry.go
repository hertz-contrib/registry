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
	"sync"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/go-redis/redis/v8"
)

var _ registry.Registry = (*redisRegistry)(nil)

type redisRegistry struct {
	client *redis.Client
	rctx   *registryContext
	mu     sync.Mutex
	wg     sync.WaitGroup
}

type registryContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// NewRedisRegistry creates a redis registry
func NewRedisRegistry(addr string, opts ...Option) registry.Registry {
	redisOpts := &redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	}
	for _, opt := range opts {
		opt(redisOpts)
	}
	rdb := redis.NewClient(redisOpts)
	return &redisRegistry{
		client: rdb,
	}
}

func (r *redisRegistry) Register(info *registry.Info) error {
	if err := validateRegistryInfo(info); err != nil {
		return err
	}
	rctx := registryContext{}
	rctx.ctx, rctx.cancel = context.WithCancel(context.Background())
	m := newMentor()
	r.wg.Add(1)
	go m.subscribe(rctx.ctx, info, r)
	r.wg.Wait()
	rdb := r.client
	hash, err := prepareRegistryHash(info)
	if err != nil {
		return err
	}
	r.mu.Lock()
	r.rctx = &rctx
	rdb.HSet(rctx.ctx, hash.key, hash.field, hash.value)
	rdb.Expire(rctx.ctx, hash.key, defaultExpireTime)
	rdb.Publish(rctx.ctx, hash.key, generateMsg(register, info.ServiceName, info.Addr.String()))
	r.mu.Unlock()
	go m.monitorTTL(rctx.ctx, hash, info, r)
	go keepAlive(rctx.ctx, hash, r)
	return nil
}

func (r *redisRegistry) Deregister(info *registry.Info) error {
	if err := validateRegistryInfo(info); err != nil {
		return err
	}
	rctx := r.rctx
	rdb := r.client
	hash, err := prepareRegistryHash(info)
	if err != nil {
		return err
	}
	r.mu.Lock()
	rdb.HDel(rctx.ctx, hash.key, hash.field)
	rdb.Publish(rctx.ctx, hash.key, generateMsg(deregister, info.ServiceName, info.Addr.String()))
	rctx.cancel()
	r.mu.Unlock()
	return nil
}
