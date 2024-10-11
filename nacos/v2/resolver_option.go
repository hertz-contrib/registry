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

import cwOption "github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/options"

type resolverOptions struct {
	cfgs []cwOption.ResolverOption
}

// ResolverOption Option is nacos registry option.
type ResolverOption func(o *resolverOptions)

// WithResolverCluster with cluster option.
func WithResolverCluster(cluster string) ResolverOption {
	return func(o *resolverOptions) {
		o.cfgs = append(o.cfgs, cwOption.WithResolverCluster(cluster))
	}
}

// WithResolverGroup with group option.
func WithResolverGroup(group string) ResolverOption {
	return func(o *resolverOptions) {
		o.cfgs = append(o.cfgs, cwOption.WithResolverGroup(group))
	}
}
