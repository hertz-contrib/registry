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
	"context"
	"github.com/cloudwego-contrib/cwgo-pkg/registry/eureka/eurekahertz"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/hudl/fargo"
)

var _ discovery.Resolver = (*eurekaResolver)(nil)

// eurekaResolver is a resolver using eureka.
type eurekaResolver struct {
	resolver discovery.Resolver
}

// NewEurekaResolver creates a eureka resolver with a slice of server addresses.
func NewEurekaResolver(servers []string) *eurekaResolver {
	return &eurekaResolver{resolver: eurekahertz.NewEurekaResolver(servers)}
}

// NewEurekaResolverFromConfig creates a eureka resolver with given configuration.
func NewEurekaResolverFromConfig(config fargo.Config) *eurekaResolver {
	return &eurekaResolver{
		resolver: eurekahertz.NewEurekaResolverFromConfig(config),
	}
}

// NewEurekaResolverFromConn creates a eureka resolver using an existing connection.
func NewEurekaResolverFromConn(conn fargo.EurekaConnection) *eurekaResolver {
	return &eurekaResolver{
		resolver: eurekahertz.NewEurekaResolverFromConn(conn),
	}
}

// Target implements the Resolver interface.
func (r *eurekaResolver) Target(ctx context.Context, target *discovery.TargetInfo) string {
	return r.resolver.Target(ctx, target)
}

// Resolve implements the Resolver interface.
func (r *eurekaResolver) Resolve(ctx context.Context, desc string) (discovery.Result, error) {
	return r.resolver.Resolve(ctx, desc)
}

// Name implements the Resolver interface.
func (r *eurekaResolver) Name() string {
	return r.resolver.Name()
}
