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

package nacos

import (
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"

	cwNacos "github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoshertz"
	cwOption "github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/options"
)

var _ registry.Registry = (*nacosRegistry)(nil)

type (
	nacosRegistry struct {
		registry registry.Registry
	}

	registryOptions struct {
		cfgs []cwOption.Option
	}

	// RegistryOption Option is nacos registry option.
	RegistryOption func(o *registryOptions)
)

// WithRegistryCluster with cluster option.
func WithRegistryCluster(cluster string) RegistryOption {
	return func(o *registryOptions) {
		o.cfgs = append(o.cfgs, cwOption.WithCluster(cluster))
	}
}

// WithRegistryGroup with group option.
func WithRegistryGroup(group string) RegistryOption {
	return func(o *registryOptions) {
		o.cfgs = append(o.cfgs, cwOption.WithGroup(group))
	}
}

func (n *nacosRegistry) Register(info *registry.Info) error {
	return n.registry.Register(info)
}

func (n *nacosRegistry) Deregister(info *registry.Info) error {
	return n.registry.Deregister(info)
}

// NewDefaultNacosRegistry create a default service registry using nacos.
func NewDefaultNacosRegistry(opts ...RegistryOption) (registry.Registry, error) {
	cfgs := transferRegistryOptions(opts...)

	nacosRegistry, err := cwNacos.NewDefaultNacosRegistry(cfgs...)
	if err != nil {
		return nil, err
	}

	return nacosRegistry, nil
}

// NewNacosRegistry create a new registry using nacos.
func NewNacosRegistry(client naming_client.INamingClient, opts ...RegistryOption) registry.Registry {
	cfgs := transferRegistryOptions(opts...)

	return &nacosRegistry{registry: cwNacos.NewNacosRegistry(client, cfgs...)}
}

// transferRegistryOptions transfer registry options to options in cwgo-pkg.
func transferRegistryOptions(opts ...RegistryOption) []cwOption.Option {
	o := &registryOptions{}

	for _, opt := range opts {
		opt(o)
	}

	return o.cfgs
}
