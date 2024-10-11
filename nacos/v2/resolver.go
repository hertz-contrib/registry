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
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"

	cwNacos "github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoshertz/v2"
	cwOption "github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/options"
)

var _ discovery.Resolver

type nacosResolver struct {
	resolver discovery.Resolver
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
	return &nacosResolver{
		resolver: cwNacos.NewNacosResolver(cli, cfgs...),
	}
}

func transferResolverOption(opts ...ResolverOption) []cwOption.ResolverOption {
	o := resolverOptions{}

	for _, opt := range opts {
		opt(&o)
	}

	return o.cfgs
}
