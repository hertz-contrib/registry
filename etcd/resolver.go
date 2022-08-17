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
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"go.etcd.io/etcd/clientv3"
)

type etcdResolver struct {
	client  *clientv3.Client
	timeout time.Duration
}

func NewEtcdResolver(servers []string, sessionTimeout time.Duration) (discovery.Resolver, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   servers,
		DialTimeout: sessionTimeout,
	})
	if err != nil {
		// handle error!
		fmt.Printf("connect to etcd failed, err:%v\n", err)
		return nil, err
	}
	return &etcdResolver{cli, sessionTimeout}, nil
}

func (e *etcdResolver) Resolve(ctx context.Context, desc string) (discovery.Result, error) {
	path := desc
	if !strings.HasPrefix(path, Separator) {
		path = Separator + path
	}

	instances, err := e.getInstances(path)
	if err != nil {
		return discovery.Result{}, err
	}
	res := discovery.Result{
		CacheKey:  desc,
		Instances: instances,
	}
	return res, nil
}

func (e *etcdResolver) getInstances(desc string) ([]discovery.Instance, error) {
	// get
	instances := make([]discovery.Instance, 0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// use the etcd get method with the prefix
	resp, err := e.client.Get(ctx, desc, clientv3.WithPrefix())
	cancel()
	if err != nil {
		fmt.Printf("get from etcd failed, err:%v\n", err)
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, errors.New("not found path")
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s:%s\n", ev.Key, ev.Value)
		key := string(ev.Key)
		value := ev.Value
		sepIndex := strings.LastIndex(string(key), Separator)
		if sepIndex <= 0 {
			return nil, errors.New("find the wrong endpoint")
		}
		ep := key[sepIndex+1:]
		host, port, err := net.SplitHostPort(ep)
		if err != nil {
			return nil, err
		}
		if port == "" {
			return nil, fmt.Errorf("missing port when parse node [%s]", ep)
		}
		if host == "" {
			return nil, fmt.Errorf("missing host when parse node [%s]", ep)
		}
		en := new(RegistryEntity)

		json.Unmarshal(value, en)

		instances = append(instances, discovery.NewInstance("tcp", ep, en.Weight, en.Tags))
	}
	return instances, nil
}

func (e *etcdResolver) Name() string {
	return "etcd"
}

func (e *etcdResolver) Target(ctx context.Context, target *discovery.TargetInfo) string {
	return target.Host
}
