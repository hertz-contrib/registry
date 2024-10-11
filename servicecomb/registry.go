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
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/go-chassis/sc-client"
)

// RegistryOption is ServiceComb option.
type RegistryOption = options.Option

// WithAppId with app id option
func WithAppId(appId string) RegistryOption {
	return options.WithAppId(appId)
}

// WithRegistryVersionRule with version rule option
func WithRegistryVersionRule(versionRule string) RegistryOption {
	return options.WithVersionRule(versionRule)
}

// WithRegistryHostName with host name option
func WithRegistryHostName(hostName string) RegistryOption {
	return options.WithHostName(hostName)
}

// WithRegistryHeartbeatInterval with heart beat second
func WithRegistryHeartbeatInterval(second int32) RegistryOption {
	return options.WithHeartbeatInterval(second)
}

// NewDefaultSCRegistry create a new default ServiceComb registry
func NewDefaultSCRegistry(endPoints []string, opts ...RegistryOption) (registry.Registry, error) {
	return servicecombhertz.NewDefaultSCRegistry(endPoints, opts...)
}

// NewSCRegistry create a new ServiceComb registry
func NewSCRegistry(client *sc.Client, opts ...RegistryOption) registry.Registry {
	return servicecombhertz.NewSCRegistry(client, opts...)
}
