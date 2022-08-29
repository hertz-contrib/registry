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
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var _ registry.Registry = (*etcdRegistry)(nil)

const (
	etcdPrefix = "hertz/registry-etcd"
	ttlKey     = "HERTZ_ETCD_REGISTRY_LEASE_TTL"
)

type etcdRegistry struct {
	etcdClient *clientv3.Client
	leaseTTL   int64
	meta       *registerMeta
	mu         sync.Mutex
}

type registerMeta struct {
	leaseID clientv3.LeaseID
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewEtcdRegistry creates a etcd based registry.
func NewEtcdRegistry(endpoints []string, opts ...Option) (registry.Registry, error) {
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
	return &etcdRegistry{
		etcdClient: etcdClient,
		leaseTTL:   getTTL(),
	}, nil
}

func (e *etcdRegistry) Register(info *registry.Info) error {
	if err := validateRegistryInfo(info); err != nil {
		return err
	}
	leaseID, err := e.grantLease()
	if err != nil {
		return err
	}

	if err := e.register(info, leaseID); err != nil {
		return err
	}
	meta := registerMeta{
		leaseID: leaseID,
	}
	meta.ctx, meta.cancel = context.WithCancel(context.Background())
	if err := e.keepalive(&meta); err != nil {
		return err
	}
	e.mu.Lock()
	e.meta = &meta
	e.mu.Unlock()
	return nil
}

func (e *etcdRegistry) Deregister(info *registry.Info) error {
	if err := validateRegistryInfo(info); err != nil {
		return err
	}
	if err := e.deregister(info); err != nil {
		return err
	}
	e.mu.Lock()
	e.meta.cancel()
	e.mu.Unlock()
	return nil
}

func (e *etcdRegistry) grantLease() (clientv3.LeaseID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()
	resp, err := e.etcdClient.Grant(ctx, e.leaseTTL)
	if err != nil {
		return clientv3.NoLease, err
	}
	return resp.ID, nil
}

func (e *etcdRegistry) register(info *registry.Info, leaseID clientv3.LeaseID) error {
	val, err := json.Marshal(&instanceInfo{
		Network: info.Addr.Network(),
		Address: info.Addr.String(),
		Weight:  info.Weight,
		Tags:    info.Tags,
	})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, err = e.etcdClient.Put(ctx, serviceKey(info.ServiceName, info.Addr.String()), string(val), clientv3.WithLease(leaseID))
	return err
}

func (e *etcdRegistry) deregister(info *registry.Info) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, err := e.etcdClient.Delete(ctx, serviceKey(info.ServiceName, info.Addr.String()))
	return err
}

// keepalive keep the lease alive
func (e *etcdRegistry) keepalive(meta *registerMeta) error {
	keepAlive, err := e.etcdClient.KeepAlive(meta.ctx, meta.leaseID)
	if err != nil {
		return err
	}
	go func() {
		// eat keepAlive channel to keep related lease alive.
		hlog.Infof("HERTZ: Start keepalive lease %x for etcd registry", meta.leaseID)
		for range keepAlive {
			select {
			case <-meta.ctx.Done():
				hlog.Infof("HERTZ: Stop keepalive lease %x for etcd registry", meta.leaseID)
				return
			default:
			}
		}
	}()
	return nil
}
