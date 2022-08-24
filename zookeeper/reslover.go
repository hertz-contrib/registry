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

package zookeeper

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/go-zookeeper/zk"
)

type zookeeperResolver struct {
	conn *zk.Conn
}

// NewZookeeperResolver create a zookeeper based resolver
func NewZookeeperResolver(servers []string, sessionTimeout time.Duration) (discovery.Resolver, error) {
	conn, _, err := zk.Connect(servers, sessionTimeout)
	if err != nil {
		return nil, err
	}
	return &zookeeperResolver{conn: conn}, nil
}

// NewZookeeperResolver create a zookeeper based resolver with auth
func NewZookeeperResolverWithAuth(servers []string, sessionTimeout time.Duration, user, password string) (discovery.Resolver, error) {
	conn, _, err := zk.Connect(servers, sessionTimeout)
	if err != nil {
		return nil, err
	}
	auth := []byte(fmt.Sprintf("%s:%s", user, password))
	err = conn.AddAuth(Scheme, auth)
	if err != nil {
		return nil, err
	}
	return &zookeeperResolver{conn: conn}, nil
}

func (z *zookeeperResolver) Target(ctx context.Context, target *discovery.TargetInfo) string {
	return target.Host
}

func (z *zookeeperResolver) Resolve(ctx context.Context, desc string) (discovery.Result, error) {
	path := desc
	if !strings.HasPrefix(path, Separator) {
		path = Separator + path
	}
	eps, err := z.getEndPoints(path)
	if err != nil {
		return discovery.Result{}, err
	}
	if len(eps) == 0 {
		return discovery.Result{CacheKey: desc}, nil
	}
	instances, err := z.getInstances(eps, path)
	if err != nil {
		return discovery.Result{}, err
	}
	res := discovery.Result{
		CacheKey:  desc,
		Instances: instances,
	}
	return res, nil
}

func (z *zookeeperResolver) getEndPoints(path string) ([]string, error) {
	child, _, err := z.conn.Children(path)
	return child, err
}

func (z *zookeeperResolver) getInstances(eps []string, path string) ([]discovery.Instance, error) {
	instances := make([]discovery.Instance, 0, len(eps))
	for _, ep := range eps {
		ins, err := z.detailEndPoints(path, ep)
		if err != nil {
			return []discovery.Instance{}, fmt.Errorf("detail endpoint [%s] info error, cause %w", ep, err)
		}
		instances = append(instances, ins)
	}
	return instances, nil
}

func (z *zookeeperResolver) detailEndPoints(path, ep string) (discovery.Instance, error) {
	data, _, err := z.conn.Get(path + Separator + ep)
	if err != nil {
		return nil, err
	}
	en := new(RegistryEntity)
	err = json.Unmarshal(data, en)
	if err != nil {
		return nil, fmt.Errorf("unmarshal data [%s] error, cause %w", data, err)
	}
	return discovery.NewInstance("tcp", ep, en.Weight, en.Tags), nil
}

func (z *zookeeperResolver) Name() string {
	return "zookeeper"
}
