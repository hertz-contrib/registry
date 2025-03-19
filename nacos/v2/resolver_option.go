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

import "github.com/nacos-group/nacos-sdk-go/v2/common/constant"

type resolverOptions struct {
	cluster      string
	group        string
	serverConfig []constant.ServerConfig
}

// ResolverOption Option is nacos registry option.
type ResolverOption func(o *resolverOptions)

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

// WithNacosServersConfig with nacos server config option.
func WithNacosServersConfig(serverConfig []constant.ServerConfig) ResolverOption {
	return func(o *resolverOptions) {
		o.serverConfig = serverConfig
	}
}
