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

package servicecomb

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/sc-client"
	"github.com/thoas/go-funk"
)

var _ registry.Registry = (*serviceCombRegistry)(nil)

type scHeartbeat struct {
	cancel      context.CancelFunc
	instanceKey string
}

type registryOptions struct {
	appId             string
	versionRule       string
	hostName          string
	heartbeatInterval int32
}

// RegistryOption is ServiceComb option.
type RegistryOption func(o *registryOptions)

// WithAppId with app id option
func WithAppId(appId string) RegistryOption {
	return func(o *registryOptions) {
		o.appId = appId
	}
}

// WithRegistryVersionRule with version rule option
func WithRegistryVersionRule(versionRule string) RegistryOption {
	return func(o *registryOptions) {
		o.versionRule = versionRule
	}
}

// WithRegistryHostName with host name option
func WithRegistryHostName(hostName string) RegistryOption {
	return func(o *registryOptions) {
		o.hostName = hostName
	}
}

// WithRegistryHeartbeatInterval with heart beat second
func WithRegistryHeartbeatInterval(second int32) RegistryOption {
	return func(o *registryOptions) {
		o.heartbeatInterval = second
	}
}

type serviceCombRegistry struct {
	cli         *sc.Client
	opts        registryOptions
	lock        *sync.RWMutex
	registryIns map[string]*scHeartbeat
}

// NewDefaultSCRegistry create a new default ServiceComb registry
func NewDefaultSCRegistry(endPoints []string, opts ...RegistryOption) (registry.Registry, error) {
	client, err := sc.NewClient(sc.Options{
		Endpoints: endPoints,
	})
	if err != nil {
		return nil, err
	}
	return NewSCRegistry(client, opts...), nil
}

// NewSCRegistry create a new ServiceComb registry
func NewSCRegistry(client *sc.Client, opts ...RegistryOption) registry.Registry {
	op := registryOptions{
		appId:             "DEFAULT",
		versionRule:       "1.0.0",
		hostName:          "DEFAULT",
		heartbeatInterval: 5,
	}
	for _, opt := range opts {
		opt(&op)
	}
	return &serviceCombRegistry{
		cli:         client,
		opts:        op,
		lock:        &sync.RWMutex{},
		registryIns: make(map[string]*scHeartbeat),
	}
}

// Register a service info to ServiceComb
func (scr *serviceCombRegistry) Register(info *registry.Info) error {
	err := scr.vaildRegistryInfo(info)
	if err != nil {
		return err
	}
	addr, err := scr.parseAddr(info.Addr.String())
	if err != nil {
		return err
	}
	instanceKey := fmt.Sprintf("%s:%s", info.ServiceName, addr)
	scr.lock.RLock()
	_, ok := scr.registryIns[instanceKey]
	scr.lock.RUnlock()
	if ok {
		return fmt.Errorf("instance{%s} already registered", instanceKey)
	}

	serviceID, err := scr.cli.RegisterService(&discovery.MicroService{
		ServiceName: info.ServiceName,
		AppId:       scr.opts.appId,
		Version:     scr.opts.versionRule,
		Status:      sc.MSInstanceUP,
	})
	if err != nil {
		return fmt.Errorf("register service error: %w", err)
	}

	healthCheck := &discovery.HealthCheck{
		Mode:     "push",
		Interval: 30,
		Times:    3,
	}
	if scr.opts.heartbeatInterval > 0 {
		healthCheck.Interval = scr.opts.heartbeatInterval
	}

	instanceId, err := scr.cli.RegisterMicroServiceInstance(&discovery.MicroServiceInstance{
		ServiceId:   serviceID,
		Endpoints:   []string{addr},
		HostName:    scr.opts.hostName,
		HealthCheck: healthCheck,
		Status:      sc.MSInstanceUP,
		Properties:  info.Tags,
	})
	if err != nil {
		return fmt.Errorf("register service instance error: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go scr.heartBeat(ctx, serviceID, instanceId)

	scr.lock.Lock()
	defer scr.lock.Unlock()
	scr.registryIns[instanceKey] = &scHeartbeat{
		instanceKey: instanceKey,
		cancel:      cancel,
	}

	return nil
}

// Deregister a service or an instance
func (scr *serviceCombRegistry) Deregister(info *registry.Info) error {
	err := scr.vaildRegistryInfo(info)
	if err != nil {
		return err
	}

	addr, err := scr.parseAddr(info.Addr.String())
	if err != nil {
		return err
	}

	serviceId, err := scr.cli.GetMicroServiceID(scr.opts.appId, info.ServiceName, scr.opts.versionRule, "")
	if err != nil {
		return fmt.Errorf("get service-id error: %w", err)
	}

	instanceKey := fmt.Sprintf("%s:%s", info.ServiceName, addr)
	scr.lock.RLock()
	insHeartbeat, ok := scr.registryIns[instanceKey]
	scr.lock.RUnlock()
	if !ok {
		return fmt.Errorf("instance{%s} has not registered", instanceKey)
	}

	instanceId := ""
	instances, err := scr.cli.FindMicroServiceInstances("", scr.opts.appId, info.ServiceName, scr.opts.versionRule, sc.WithoutRevision())
	if err != nil {
		return fmt.Errorf("get instances error: %w", err)
	}
	for _, instance := range instances {
		if funk.ContainsString(instance.Endpoints, addr) {
			instanceId = instance.InstanceId
		}
	}
	if instanceId != "" {
		// unregister is to slow to effect, mark it down first.
		_, err = scr.cli.UpdateMicroServiceInstanceStatus(serviceId, instanceId, sc.MSIinstanceDown)
		if err != nil {
			return fmt.Errorf("down service error: %w", err)
		}
		_, err = scr.cli.UnregisterMicroServiceInstance(serviceId, instanceId)
		if err != nil {
			return fmt.Errorf("deregister service error: %w", err)
		}
	}

	scr.lock.Lock()
	insHeartbeat.cancel()
	delete(scr.registryIns, instanceKey)
	scr.lock.Unlock()
	return nil
}

func (scr *serviceCombRegistry) heartBeat(ctx context.Context, serviceId, instanceId string) {
	ticker := time.NewTicker(time.Second * time.Duration(scr.opts.heartbeatInterval))
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			success, err := scr.cli.Heartbeat(serviceId, instanceId)
			if err != nil || !success {
				hlog.CtxErrorf(ctx, "HERTZ: beat to ServerComb return error:%+v instance:%v", err, instanceId)
				ticker.Stop()
				return
			}
		}
	}
}

func (scr *serviceCombRegistry) vaildRegistryInfo(info *registry.Info) error {
	if info == nil {
		return errors.New("registry.Info can not be empty")
	}
	if info.ServiceName == "" {
		return errors.New("registry.Info ServiceName can not be empty")
	}
	if info.Addr == nil {
		return errors.New("registry.Info Addr can not be empty")
	}
	return nil
}

func (scr *serviceCombRegistry) parseAddr(s string) (string, error) {
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return "", fmt.Errorf("parse deregistry info addr error: %w", err)
	}
	if host == "" || host == "::" {
		host = utils.LocalIP()
	}

	return net.JoinHostPort(host, port), nil
}
