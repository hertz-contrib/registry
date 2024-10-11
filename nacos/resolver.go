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
	"context"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"

	cwNacos "github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoshertz"
	cwOption "github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/options"
)

var _ discovery.Resolver = (*nacosResolver)(nil)

type (
	resolverOptions struct {
		cgfs []cwOption.ResolverOption
	}

	// ResolverOption Option is nacos registry option.
	ResolverOption func(o *resolverOptions)

	nacosResolver struct {
		resolver discovery.Resolver
	}
)

// WithResolverCluster with cluster option.
func WithResolverCluster(cluster string) ResolverOption {
	return func(o *resolverOptions) {
		o.cgfs = append(o.cgfs, cwOption.WithResolverCluster(cluster))
	}
}

// WithResolverGroup with group option.
func WithResolverGroup(group string) ResolverOption {
	return func(o *resolverOptions) {
		o.cgfs = append(o.cgfs, cwOption.WithResolverGroup(group))
	}
}

func (n *nacosResolver) Target(ctx context.Context, target *discovery.TargetInfo) string {
	return n.resolver.Target(ctx, target)
}

func (n *nacosResolver) Resolve(ctx context.Context, desc string) (discovery.Result, error) {
	return n.resolver.Resolve(ctx, desc)
}

func (n *nacosResolver) Name() string {
	return n.resolver.Name()
}

// NewDefaultNacosResolver create a default service resolver using nacos.
func NewDefaultNacosResolver(opts ...ResolverOption) (discovery.Resolver, error) {
	cfgs := transferResolverOption(opts...)

	nacosResolver, err := cwNacos.NewDefaultNacosResolver(cfgs...)
	if err != nil {
		return nil, err
	}

	return nacosResolver, nil
}

// NewNacosResolver create a service resolver using nacos.
func NewNacosResolver(cli naming_client.INamingClient, opts ...ResolverOption) discovery.Resolver {
	cfgs := transferResolverOption(opts...)

	return &nacosResolver{resolver: cwNacos.NewNacosResolver(cli, cfgs...)}
}

// transferResolverOption transfer local ResolverOption to ResolverOption in cwgo-pkg.
func transferResolverOption(opts ...ResolverOption) []cwOption.ResolverOption {
	o := &resolverOptions{}

	for _, opt := range opts {
		opt(o)
	}

	return o.cgfs
}

// compareMaps compares two maps regardless of nil or empty
func compareMaps(m1, m2 map[string]string) bool {
	// if both maps are nil, they are equal
	if m1 == nil && m2 == nil {
		return true
	}
	// if the lengths are different, the maps are not equal
	if len(m1) != len(m2) {
		return false
	}
	// iterate over the keys of m1 and check if they exist in m2 with the same value
	for k, v := range m1 {
		if v2, ok := m2[k]; !ok || v != v2 {
			return false
		}
	}
	// return true if no differences are found
	return true
}
