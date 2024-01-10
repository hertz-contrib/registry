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

type registryOptions struct {
	cluster string
	group   string
}

// RegistryOption Option is nacos registry option.
type RegistryOption func(o *registryOptions)

// WithRegistryCluster with cluster option.
func WithRegistryCluster(cluster string) RegistryOption {
	return func(o *registryOptions) {
		o.cluster = cluster
	}
}

// WithRegistryGroup with group option.
func WithRegistryGroup(group string) RegistryOption {
	return func(o *registryOptions) {
		o.group = group
	}
}
