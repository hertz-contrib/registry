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
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var _ registry.Registry = (*etcdRegistry)(nil)

const (
	etcdPrefix          = "hertz/registry-etcd"
	ttlKey              = "HERTZ_ETCD_REGISTRY_LEASE_TTL"
	hertzIpToRegistry   = "HERTZ_IP_TO_REGISTRY"
	hertzPortToRegistry = "HERTZ_PORT_TO_REGISTRY"
)

type etcdRegistry struct {
	etcdClient  *clientv3.Client
	retryConfig *retryCfg

	leaseTTL int64
	meta     *registerMeta
	mu       sync.Mutex
	stop     chan struct{}
}

type registerMeta struct {
	leaseID clientv3.LeaseID
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewEtcdRegistry creates a etcd based registry.
func NewEtcdRegistry(endpoints []string, opts ...Option) (registry.Registry, error) {
	cfg := &option{
		etcdCfg: clientv3.Config{
			Endpoints: endpoints,
		},
		retryCfg: &retryCfg{
			maxAttemptTimes: 5,
			observeDelay:    30 * time.Second,
			retryDelay:      10 * time.Second,
		},
	}
	cfg.apply(opts...)

	etcdClient, err := clientv3.New(cfg.etcdCfg)
	if err != nil {
		return nil, err
	}
	return &etcdRegistry{
		etcdClient:  etcdClient,
		leaseTTL:    getTTL(),
		retryConfig: cfg.retryCfg,
		stop:        make(chan struct{}, 1),
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
	if err := e.keepalive(meta); err != nil {
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	resp, err := e.etcdClient.Grant(ctx, e.leaseTTL)
	if err != nil {
		return clientv3.NoLease, err
	}
	return resp.ID, nil
}

func (e *etcdRegistry) register(info *registry.Info, leaseID clientv3.LeaseID) error {
	addr, err := e.getAddressOfRegistration(info)
	if err != nil {
		return err
	}
	val, err := sonic.Marshal(&instanceInfo{
		Network: info.Addr.Network(),
		Address: addr,
		Weight:  info.Weight,
		Tags:    info.Tags,
	})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, err = e.etcdClient.Put(ctx, serviceKey(info.ServiceName, addr), string(val), clientv3.WithLease(leaseID))
	if err != nil {
		return err
	}

	// retry start
	go func(key, val string) {
		e.keepRegister(key, val, e.retryConfig)
	}(serviceKey(info.ServiceName, addr), string(val))

	return nil
}

func (e *etcdRegistry) deregister(info *registry.Info) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	addr, err := e.getAddressOfRegistration(info)
	if err != nil {
		return err
	}
	_, err = e.etcdClient.Delete(ctx, serviceKey(info.ServiceName, addr))
	if err != nil {
		return err
	}
	e.stop <- struct{}{}
	return nil
}

// keepalive keep the lease alive
func (e *etcdRegistry) keepalive(meta registerMeta) error {
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

// keepRegister keep register service by retryConfig
func (e *etcdRegistry) keepRegister(key, val string, retryConfig *retryCfg) {
	var (
		failedTimes uint
		resp        *clientv3.GetResponse
		err         error
		ctx         context.Context
		cancel      context.CancelFunc
		wg          sync.WaitGroup
	)

	delay := retryConfig.observeDelay
	// if maxAttemptTimes is 0, keep register forever
	for retryConfig.maxAttemptTimes == 0 || failedTimes < retryConfig.maxAttemptTimes {
		select {
		case _, ok := <-e.stop:
			if !ok {
				close(e.stop)
			}
			hlog.Infof("stop keep register service %s", key)
			return
		case <-time.After(delay):
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel = context.WithTimeout(context.Background(), time.Second*3)
			resp, err = e.etcdClient.Get(ctx, key)
			defer cancel()
		}()
		wg.Wait()

		if err != nil {
			hlog.Warnf("keep register get %s failed with err: %v", key, err)
			delay = retryConfig.retryDelay
			failedTimes++
			continue
		}

		if len(resp.Kvs) == 0 {
			hlog.Infof("keep register service %s", key)
			delay = retryConfig.retryDelay
			leaseID, err := e.grantLease()
			if err != nil {
				hlog.Warnf("keep register grant lease %s failed with err: %v", key, err)
				failedTimes++
				continue
			}

			_, err = e.etcdClient.Put(ctx, key, val, clientv3.WithLease(leaseID))
			if err != nil {
				hlog.Warnf("keep register put %s failed with err: %v", key, err)
				failedTimes++
				continue
			}

			meta := registerMeta{
				leaseID: leaseID,
			}
			meta.ctx, meta.cancel = context.WithCancel(context.Background())
			if err := e.keepalive(meta); err != nil {
				hlog.Warnf("keep register keepalive %s failed with err: %v", key, err)
				failedTimes++
				continue
			}
			e.meta.cancel()
			e.meta = &meta
			delay = retryConfig.observeDelay
		}
		failedTimes = 0
	}
	hlog.Errorf("keep register service %s failed times:%d", key, failedTimes)
}

// getAddressOfRegistration returns the address of the service registration.
func (e *etcdRegistry) getAddressOfRegistration(info *registry.Info) (string, error) {
	host, port, err := net.SplitHostPort(info.Addr.String())
	if err != nil {
		return "", err
	}

	// if host is empty or "::", use local ipv4 address as host
	if host == "" || host == "::" {
		host, err = getLocalIPv4Host()
		if err != nil {
			return "", fmt.Errorf("parse registry info addr error: %w", err)
		}
	}

	// if env HERTZ_IP_TO_REGISTRY is set, use it as host
	if ipToRegistry, exists := os.LookupEnv(hertzIpToRegistry); exists && ipToRegistry != "" {
		host = ipToRegistry
	}

	// if env HERTZ_PORT_TO_REGISTRY is set, use it as port
	if portToRegistry, exists := os.LookupEnv(hertzPortToRegistry); exists && portToRegistry != "" {
		port = portToRegistry
	}

	return net.JoinHostPort(host, port), nil
}

func getLocalIPv4Host() (string, error) {
	addr, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addr {
		ipNet, isIpNet := addr.(*net.IPNet)
		if isIpNet && !ipNet.IP.IsLoopback() {
			ipv4 := ipNet.IP.To4()
			if ipv4 != nil {
				return ipv4.String(), nil
			}
		}
	}
	return "", fmt.Errorf("not found ipv4 address")
}
