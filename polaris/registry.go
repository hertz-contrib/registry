/*
 * Copyright 2021 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package polaris

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/model"
)

var (
	defaultHeartbeatIntervalSec = 5
	registerTimeout             = 10 * time.Second
	heartbeatTimeout            = 5 * time.Second
	heartbeatTime               = 5 * time.Second
)

// Registry is extension interface of Hertz registry.Registry.
type Registry interface {
	registry.Registry

	doHeartbeat(ctx context.Context, ins *api.InstanceRegisterRequest)
}

type polarisHeartbeat struct {
	cancel      context.CancelFunc
	instanceKey string
}

// polarisRegistry is a registry using polaris.
type polarisRegistry struct {
	consumer    api.ConsumerAPI
	provider    api.ProviderAPI
	lock        *sync.RWMutex
	registryIns map[string]*polarisHeartbeat
}

// NewPolarisRegistry creates a polaris based registry.
func NewPolarisRegistry(configFile ...string) (Registry, error) {
	sdkCtx, err := GetPolarisConfig(configFile...)
	if err != nil {
		return nil, err
	}

	pRegistry := &polarisRegistry{
		consumer:    api.NewConsumerAPIByContext(sdkCtx),
		provider:    api.NewProviderAPIByContext(sdkCtx),
		registryIns: make(map[string]*polarisHeartbeat),
		lock:        &sync.RWMutex{},
	}

	return pRegistry, nil
}

// Register registers a server with given registry info.
func (svr *polarisRegistry) Register(info *registry.Info) error {
	if err := validateInfo(info); err != nil {
		return err
	}
	param, instanceKey, err := createRegisterParam(info)
	if err != nil {
		return err
	}
	resp, err := svr.provider.Register(param)
	if err != nil {
		return err
	}
	if resp.Existed {
		hlog.Warnf("HERTZ: instance already registered, namespace:%s, service:%s, port:%s",
			param.Namespace, param.Service, param.Host)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go svr.doHeartbeat(ctx, param)
	svr.lock.Lock()
	defer svr.lock.Unlock()
	svr.registryIns[instanceKey] = &polarisHeartbeat{
		instanceKey: instanceKey,
		cancel:      cancel,
	}
	return nil
}

// Deregister deregisters a server with given registry info.
func (svr *polarisRegistry) Deregister(info *registry.Info) error {
	if err := validateInfo(info); err != nil {
		return err
	}
	request, instanceKey, err := createDeregisterParam(info)
	if err != nil {
		return err
	}
	svr.lock.RLock()
	insHeartbeat, ok := svr.registryIns[instanceKey]
	svr.lock.RUnlock()
	if !ok {
		return fmt.Errorf("instance{%s} has not registered", instanceKey)
	}
	err = svr.provider.Deregister(request)
	if err != nil {
		return fmt.Errorf("instance{%s} deregister fail (err:%+v)", instanceKey, err)
	}

	svr.lock.Lock()
	insHeartbeat.cancel()
	delete(svr.registryIns, instanceKey)
	svr.lock.Unlock()

	return nil
}

// IsAvailable always return true when use polaris.
func (svr *polarisRegistry) IsAvailable() bool {
	return true
}

// doHeartbeat Since polaris does not support automatic reporting of instance heartbeats, separate logic is needed to implement it.
func (svr *polarisRegistry) doHeartbeat(ctx context.Context, ins *api.InstanceRegisterRequest) {
	ticker := time.NewTicker(heartbeatTime)

	heartbeat := &api.InstanceHeartbeatRequest{
		InstanceHeartbeatRequest: model.InstanceHeartbeatRequest{
			Service:   ins.Service,
			Namespace: ins.Namespace,
			Host:      ins.Host,
			Port:      ins.Port,
			Timeout:   model.ToDurationPtr(heartbeatTimeout),
		},
	}
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			if err := svr.provider.Heartbeat(heartbeat); err != nil {
				hlog.Errorf("HERTZ: Heartbeat error, reason: %s", err.Error())
			}
		}
	}
}

// validateInfo validates registry.Info.
func validateInfo(info *registry.Info) error {
	if info.ServiceName == "" {
		return fmt.Errorf("missing service name in Register")
	}
	if info.Addr == nil {
		return fmt.Errorf("missing Addr in Register")
	}
	if len(info.Addr.Network()) == 0 {
		return fmt.Errorf("registry.Info Addr Network() can not be empty")
	}
	if len(info.Addr.String()) == 0 {
		return fmt.Errorf("registry.Info Addr String() can not be empty")
	}
	return nil
}

// createRegisterParam convert registry.Info to polaris instance register request.
func createRegisterParam(info *registry.Info) (*api.InstanceRegisterRequest, string, error) {
	instanceHost, instancePort, err := GetInfoHostAndPort(info.Addr.String())
	if err != nil {
		return nil, "", err
	}
	protocol := info.Addr.Network()

	namespace, ok := info.Tags["namespace"]
	if !ok {
		namespace = polarisDefaultNamespace
	}
	instanceKey := GetInstanceKey(namespace, info.ServiceName, instanceHost, strconv.Itoa(instancePort))

	req := &api.InstanceRegisterRequest{
		InstanceRegisterRequest: model.InstanceRegisterRequest{
			Service:   info.ServiceName,
			Namespace: namespace,
			Host:      instanceHost,
			Port:      instancePort,
			Protocol:  &protocol,
			Timeout:   model.ToDurationPtr(registerTimeout),
			TTL:       &defaultHeartbeatIntervalSec,
			// If the TTL field is not set, polaris will think that this instance does not need to perform the heartbeat health check operation,
			// then after the instance goes offline, the instance cannot be converted to unhealthy normally.
		},
	}

	return req, instanceKey, nil
}

// createDeregisterParam convert registry.info to polaris instance deregister request.
func createDeregisterParam(info *registry.Info) (*api.InstanceDeRegisterRequest, string, error) {
	instanceHost, instancePort, err := GetInfoHostAndPort(info.Addr.String())
	if err != nil {
		return nil, "", err
	}

	namespace, ok := info.Tags["namespace"]
	if !ok {
		namespace = polarisDefaultNamespace
	}

	instanceKey := GetInstanceKey(namespace, info.ServiceName, instanceHost, strconv.Itoa(instancePort))
	req := &api.InstanceDeRegisterRequest{
		InstanceDeRegisterRequest: model.InstanceDeRegisterRequest{
			Service:   info.ServiceName,
			Namespace: namespace,
			Host:      instanceHost,
			Port:      instancePort,
		},
	}
	return req, instanceKey, nil
}
