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
	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"

	cwNacos "github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoshertz/v2"
)

// NewDefaultNacosResolver create a default service resolver using nacos.
func NewDefaultNacosResolver(opts ...ResolverOption) (discovery.Resolver, error) {
	return cwNacos.NewDefaultNacosResolver(opts...)
}

// NewNacosResolver create a service resolver using nacos.
func NewNacosResolver(cli naming_client.INamingClient, opts ...ResolverOption) discovery.Resolver {
	return cwNacos.NewNacosResolver(cli, opts...)
}
