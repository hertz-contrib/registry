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

import nacosOption "github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/options"

// ResolverOption Option is nacos registry option.
type ResolverOption = nacosOption.ResolverOption

// WithResolverCluster with cluster option.
func WithResolverCluster(cluster string) ResolverOption {
	return nacosOption.WithResolverCluster(cluster)
}

// WithResolverGroup with group option.
func WithResolverGroup(group string) ResolverOption {
	return nacosOption.WithResolverGroup(group)
}
