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
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/go-redis/redis/v8"
)

var _ discovery.Resolver = (*redisResolver)(nil)

type redisResolver struct {
	client *redis.Client
}

// NewRedisResolver create a redis resolver
func NewRedisResolver(addr string, opts ...Option) discovery.Resolver {
	redisOpts := &redis.Options{Addr: addr}
	for _, opt := range opts {
		opt(redisOpts)
	}
	rdb := redis.NewClient(redisOpts)
	return &redisResolver{
		client: rdb,
	}
}

func (r *redisResolver) Target(_ context.Context, target *discovery.TargetInfo) string {
	return target.Host
}

func (r *redisResolver) Resolve(ctx context.Context, desc string) (discovery.Result, error) {
	rdb := r.client
	fvs := rdb.HGetAll(ctx, fmt.Sprintf("/%s/%s/%s", hertz, desc, server)).Val()
	var (
		ri  registryInfo
		its []discovery.Instance
	)
	for f, v := range fvs {
		err := json.Unmarshal([]byte(v), &ri)
		if err != nil {
			hlog.Warnf("HERTZ: fail to unmarshal with err: %v, ignore instance Addr: %v", err, f)
		}
		weight := ri.Weight
		if weight <= 0 {
			weight = defaultWeight
		}
		its = append(its, discovery.NewInstance(tcp, ri.Addr, weight, ri.Tags))
	}
	return discovery.Result{
		CacheKey:  desc,
		Instances: its,
	}, nil
}

func (r *redisResolver) Name() string {
	return Redis
}
