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

package consul

import (
	"github.com/cloudwego-contrib/cwgo-pkg/registry/consul/consulhertz"
	cwOption "github.com/cloudwego-contrib/cwgo-pkg/registry/consul/options"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/hashicorp/consul/api"
)

type consulRegistry struct {
	registry registry.Registry
}

var _ registry.Registry = (*consulRegistry)(nil)

type options struct {
	cfgs []cwOption.Option
}

// Option is the option of Consul.
type Option func(o *options)

// WithCheck is consul registry option to set AgentServiceCheck.
func WithCheck(check *api.AgentServiceCheck) Option {
	return func(o *options) {
		o.cfgs = append(o.cfgs, cwOption.WithCheck(check))
	}
}

// NewConsulRegister create a new registry using consul.
func NewConsulRegister(consulClient *api.Client, opts ...Option) registry.Registry {
	o := options{}

	for _, opt := range opts {
		opt(&o)
	}

	return &consulRegistry{registry: consulhertz.NewConsulRegister(consulClient, o.cfgs...)}
}

// Register register a service to consul.
func (c *consulRegistry) Register(info *registry.Info) error {
	return c.registry.Register(info)
}

// Deregister deregister a service from consul.
func (c *consulRegistry) Deregister(info *registry.Info) error {
	return c.registry.Deregister(info)
}
