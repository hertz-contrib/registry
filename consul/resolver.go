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
	"context"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/consul/consulhertz"
	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/hashicorp/consul/api"
)

type consulResolver struct {
	resolver discovery.Resolver
}

var _ discovery.Resolver = (*consulResolver)(nil)

// NewConsulResolver create a service resolver using consul.
func NewConsulResolver(consulClient *api.Client) discovery.Resolver {
	return &consulResolver{resolver: consulhertz.NewConsulResolver(consulClient)}
}

// Target return a description for the given target that is suitable for being a key for cache.
func (c *consulResolver) Target(ctx context.Context, target *discovery.TargetInfo) (description string) {
	return c.resolver.Target(ctx, target)
}

// Name returns the name of the resolver.
func (c *consulResolver) Name() string {
	return c.resolver.Name()
}

// Resolve a service info by desc.
func (c *consulResolver) Resolve(ctx context.Context, desc string) (discovery.Result, error) {
	return c.resolver.Resolve(ctx, desc)
}
