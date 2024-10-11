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
	"github.com/cloudwego-contrib/cwgo-pkg/registry/consul/consulhertz"
	cwOption "github.com/cloudwego-contrib/cwgo-pkg/registry/consul/options"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/hashicorp/consul/api"
)

// Option is the option of Consul.
type Option = cwOption.Option

// WithCheck is consul registry option to set AgentServiceCheck.
func WithCheck(check *api.AgentServiceCheck) Option {
	return cwOption.WithCheck(check)
}

// NewConsulRegister create a new registry using consul.
func NewConsulRegister(consulClient *api.Client, opts ...Option) registry.Registry {
	return consulhertz.NewConsulRegister(consulClient, opts...)
}
