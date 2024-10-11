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

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/polarismesh/polaris-go/api"

	cwPolaris "github.com/cloudwego-contrib/cwgo-pkg/registry/polaris/polarishertz"
)

// Registry is extension interface of Hertz registry.Registry.
type Registry interface {
	registry.Registry

	doHeartbeat(ctx context.Context, ins *api.InstanceRegisterRequest)
}

// polarisRegistry is a registry using polaris.
type polarisRegistry struct {
	registry cwPolaris.PolarisRegistry
}

// NewPolarisRegistry creates a polaris based registry.
func NewPolarisRegistry(configFile ...string) (Registry, error) {
	registry, err := cwPolaris.NewPolarisRegistry(configFile...)
	if err != nil {
		return nil, err
	}

	cwpRegistry, ok := registry.(*cwPolaris.PolarisRegistry)
	if !ok {
		return nil, fmt.Errorf("type assertion failed")
	}

	pRegistry := &polarisRegistry{
		registry: *cwpRegistry,
	}

	return pRegistry, nil
}

// Register registers a server with given registry info.
func (svr *polarisRegistry) Register(info *registry.Info) error {
	return svr.registry.Register(info)
}

// Deregister deregisters a server with given registry info.
func (svr *polarisRegistry) Deregister(info *registry.Info) error {
	return svr.registry.Deregister(info)
}

// IsAvailable always return true when use polaris.
func (svr *polarisRegistry) IsAvailable() bool {
	return svr.registry.IsAvailable()
}

// doHeartbeat Since polaris does not support automatic reporting of instance heartbeats, separate logic is needed to implement it.
func (svr *polarisRegistry) doHeartbeat(ctx context.Context, ins *api.InstanceRegisterRequest) {
	svr.registry.DoHeartbeat(ctx, ins)
}
