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

package servicecomb

import (
	"github.com/cloudwego-contrib/cwgo-pkg/registry/servicecomb/options"
	"github.com/cloudwego-contrib/cwgo-pkg/registry/servicecomb/servicecombhertz"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/go-chassis/sc-client"
)

type resolverOptions struct {
	options []options.ResolverOption
}

// ResolverOption is service-comb resolver option.
type ResolverOption = options.ResolverOption

// WithResolverAppId with appId option.
func WithResolverAppId(appId string) ResolverOption {
	return options.WithResolverAppId(appId)
}

// WithResolverVersionRule with versionRule option.
func WithResolverVersionRule(versionRule string) ResolverOption {
	return options.WithResolverVersionRule(versionRule)
}

// WithResolverConsumerId with consumerId option.
func WithResolverConsumerId(consumerId string) ResolverOption {
	return options.WithResolverConsumerId(consumerId)
}

func NewDefaultSCResolver(endPoints []string, opts ...ResolverOption) (discovery.Resolver, error) {
	return servicecombhertz.NewDefaultSCResolver(endPoints, opts...)
}

func NewSCResolver(cli *sc.Client, opts ...ResolverOption) discovery.Resolver {
	return servicecombhertz.NewSCResolver(cli, opts...)

}
