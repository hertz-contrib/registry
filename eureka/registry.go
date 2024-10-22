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

package eureka

import (
	"errors"
	"time"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/eureka/eurekahertz"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/hudl/fargo"
)

const (
	Eureka = "eureka"
	TCP    = "tcp"
	Meta   = "meta"
)

var _ registry.Registry = (*eurekaRegistry)(nil)

var (
	ErrNilInfo          = errors.New("registry info can't be nil")
	ErrNilAddr          = errors.New("registry addr can't be nil")
	ErrEmptyServiceName = errors.New("registry service name can't be empty")
	ErrMissingPort      = errors.New("addr missing port")
)

type RegistryEntity struct {
	Weight int
	Tags   map[string]string
}

type eurekaRegistry struct {
	registry registry.Registry
}

// NewEurekaRegistry creates a eureka registry.
func NewEurekaRegistry(servers []string, heatBeatInterval time.Duration) *eurekaRegistry {
	return &eurekaRegistry{
		registry: eurekahertz.NewEurekaRegistry(servers, heatBeatInterval),
	}
}

// NewEurekaRegistryFromConfig creates a eureka registry.
func NewEurekaRegistryFromConfig(config fargo.Config, heatBeatInterval time.Duration) *eurekaRegistry {
	return &eurekaRegistry{
		registry: eurekahertz.NewEurekaRegistryFromConfig(config, heatBeatInterval),
	}
}

// NewEurekaRegistryFromConn creates a eureka registry.
func NewEurekaRegistryFromConn(conn fargo.EurekaConnection, heatBeatInterval time.Duration) *eurekaRegistry {
	return &eurekaRegistry{
		registry: eurekahertz.NewEurekaRegistryFromConn(conn, heatBeatInterval),
	}
}

// Deregister deregister a server with given registry info.
func (e *eurekaRegistry) Deregister(info *registry.Info) error {
	return e.registry.Deregister(info)
}

// Register a server with given registry info.
func (e *eurekaRegistry) Register(info *registry.Info) error {
	return e.registry.Register(info)
}
