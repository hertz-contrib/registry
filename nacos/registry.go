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
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"

	nacoshertz "github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoshertz"
	nacosOption "github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/options"
)

type (
	// RegistryOption Option is nacos registry option.
	RegistryOption = nacosOption.Option
)

// WithRegistryCluster with cluster option.
func WithRegistryCluster(cluster string) RegistryOption {
	return nacosOption.WithCluster(cluster)
}

// WithRegistryGroup with group option.
func WithRegistryGroup(group string) RegistryOption {
	return nacosOption.WithGroup(group)
}

// NewDefaultNacosRegistry create a default service registry using nacos.
func NewDefaultNacosRegistry(opts ...RegistryOption) (registry.Registry, error) {
	return nacoshertz.NewDefaultNacosRegistry(opts...)
}

// NewNacosRegistry create a new registry using nacos.
func NewNacosRegistry(client naming_client.INamingClient, opts ...RegistryOption) registry.Registry {
	return nacoshertz.NewNacosRegistry(client, opts...)
}
