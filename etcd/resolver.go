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

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var _ discovery.Resolver = (*etcdResolver)(nil)

type etcdResolver struct {
	etcdClient *clientv3.Client
}

// NewEtcdResolver creates a etcd based resolver.
func NewEtcdResolver(endpoints []string, opts ...Option) (discovery.Resolver, error) {
	cfg := clientv3.Config{
		Endpoints: endpoints,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	etcdClient, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	return &etcdResolver{
		etcdClient: etcdClient,
	}, nil
}

// Resolve implements the Resolver interface.
func (e *etcdResolver) Resolve(ctx context.Context, desc string) (discovery.Result, error) {
	path := desc + "/"
	prefix := serviceKeyPrefix(path)
	resp, err := e.etcdClient.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return discovery.Result{}, err
	}
	var (
		info instanceInfo
		eps  []discovery.Instance
	)
	for _, kv := range resp.Kvs {
		err := json.Unmarshal(kv.Value, &info)
		if err != nil {
			hlog.Warnf("HERTZ: fail to unmarshal with err: %v, ignore key: %v", err, string(kv.Key))
			continue
		}
		weight := info.Weight
		if weight <= 0 {
			weight = registry.DefaultWeight
		}
		eps = append(eps, discovery.NewInstance(info.Network, info.Address, weight, info.Tags))
	}
	return discovery.Result{
		CacheKey:  desc,
		Instances: eps,
	}, nil
}

func (e *etcdResolver) Name() string {
	return "etcd"
}

func (e *etcdResolver) Target(ctx context.Context, target *discovery.TargetInfo) string {
	return target.Host
}
