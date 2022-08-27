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
	"fmt"
	"net"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/hashicorp/consul/api"
)

const (
	defaultNetwork = "tcp"
)

type consulResolver struct {
	consulClient *api.Client
}

var _ discovery.Resolver = (*consulResolver)(nil)

// NewConsulResolver create a service resolver using consul.
func NewConsulResolver(consulClient *api.Client, opts ...Option) discovery.Resolver {
	op := options{}
	for _, opt := range opts {
		opt(&op)
	}

	return &consulResolver{consulClient: consulClient}
}

// Target return a description for the given target that is suitable for being a key for cache.
func (c *consulResolver) Target(ctx context.Context, target *discovery.TargetInfo) (description string) {
	return target.Host
}

// Name returns the name of the resolver.
func (c *consulResolver) Name() string {
	return "consul"
}

// Resolve a service info by desc.
func (c *consulResolver) Resolve(ctx context.Context, desc string) (discovery.Result, error) {
	var eps []discovery.Instance
	agentServiceList, _, err := c.consulClient.Health().Service(desc, "", true, nil)
	if err != nil {
		return discovery.Result{}, err
	}
	if len(agentServiceList) == 0 {
		return discovery.Result{CacheKey: desc}, nil
	}
	for _, i := range agentServiceList {
		svc := i.Service
		if svc == nil || svc.Address == "" {
			continue
		}
		eps = append(eps, discovery.NewInstance(
			defaultNetwork,
			net.JoinHostPort(svc.Address, fmt.Sprintf("%d", svc.Port)),
			svc.Weights.Passing,
			svc.Meta,
		))
	}

	return discovery.Result{
		CacheKey:  desc,
		Instances: eps,
	}, nil
}
