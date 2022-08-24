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
	"net"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/hertz-contrib/registry/nacos/common"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

var _ discovery.Resolver = (*nacosResolver)(nil)

type (
	resolverOptions struct {
		cluster string
		group   string
	}

	// ResolverOption Option is nacos registry option.
	ResolverOption func(o *resolverOptions)

	nacosResolver struct {
		client naming_client.INamingClient
		opts   resolverOptions
	}
)

// WithResolverCluster with cluster option.
func WithResolverCluster(cluster string) ResolverOption {
	return func(o *resolverOptions) {
		o.cluster = cluster
	}
}

// WithResolverGroup with group option.
func WithResolverGroup(group string) ResolverOption {
	return func(o *resolverOptions) {
		o.group = group
	}
}

func (n *nacosResolver) Target(_ context.Context, target *discovery.TargetInfo) string {
	return target.Host
}

func (n *nacosResolver) Resolve(_ context.Context, desc string) (discovery.Result, error) {
	res, err := n.client.SelectInstances(vo.SelectInstancesParam{
		ServiceName: desc,
		HealthyOnly: true,
		GroupName:   n.opts.group,
		Clusters:    []string{n.opts.cluster},
	})
	if err != nil {
		return discovery.Result{}, err
	}

	if len(res) == 0 {
		return discovery.Result{}, nil
	}

	instances := make([]discovery.Instance, 0, len(res))
	for _, in := range res {
		if !in.Enable {
			continue
		}
		formatPort := strconv.FormatUint(in.Port, 10)
		instances = append(instances,
			discovery.NewInstance(
				"tcp",
				net.JoinHostPort(in.Ip, formatPort),
				int(in.Weight), in.Metadata,
			),
		)
	}

	return discovery.Result{
		CacheKey:  desc,
		Instances: instances,
	}, nil
}

func (n *nacosResolver) Name() string {
	return "nacos" + ":" + n.opts.cluster + ":" + n.opts.group
}

// NewDefaultNacosResolver create a default service resolver using nacos.
func NewDefaultNacosResolver(opts ...ResolverOption) (discovery.Resolver, error) {
	client, err := common.NewDefaultNacosConfig()
	if err != nil {
		return nil, err
	}
	return NewNacosResolver(client, opts...), nil
}

// NewNacosResolver create a service resolver using nacos.
func NewNacosResolver(cli naming_client.INamingClient, opts ...ResolverOption) discovery.Resolver {
	opt := resolverOptions{
		cluster: "DEFAULT",
		group:   "DEFAULT_GROUP",
	}
	for _, option := range opts {
		option(&opt)
	}
	return &nacosResolver{client: cli, opts: opt}
}
