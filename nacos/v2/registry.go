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
	"fmt"
	"net"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

var _ registry.Registry = (*nacosRegistry)(nil)

type (
	nacosRegistry struct {
		client naming_client.INamingClient
		opts   registryOptions
	}
)

// NewDefaultNacosRegistry create a default service registry using nacos.
func NewDefaultNacosRegistry(opts ...RegistryOption) (registry.Registry, error) {
	client, err := newDefaultNacosConfig()
	if err != nil {
		return nil, err
	}
	return NewNacosRegistry(client, opts...), nil
}

// NewNacosRegistry create a new registry using nacos.
func NewNacosRegistry(client naming_client.INamingClient, opts ...RegistryOption) registry.Registry {
	opt := registryOptions{
		cluster: "DEFAULT",
		group:   "DEFAULT_GROUP",
	}
	for _, option := range opts {
		option(&opt)
	}
	return &nacosRegistry{client: client, opts: opt}
}

func (n *nacosRegistry) Register(info *registry.Info) error {
	if err := n.validRegistryInfo(info); err != nil {
		return fmt.Errorf("valid parse registry info error: %w", err)
	}

	host, port, err := net.SplitHostPort(info.Addr.String())
	if err != nil {
		return fmt.Errorf("parse registry info addr error: %w", err)
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("parse registry info port error: %w", err)
	}
	ip := net.ParseIP(host)
	if ip == nil || ip.IsUnspecified() {
		host = utils.LocalIP()
	}
	success, err := n.client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          host,
		Port:        uint64(p),
		ServiceName: info.ServiceName,
		GroupName:   n.opts.group,
		ClusterName: n.opts.cluster,
		Weight:      float64(info.Weight),
		Enable:      true,
		Ephemeral:   true,
		Healthy:     true,
		Metadata:    info.Tags,
	})
	if success {
		hlog.SystemLogger().Info("register instance success")
	}
	if err != nil {
		return fmt.Errorf("register instance error: %w", err)
	}

	return nil
}

func (n *nacosRegistry) Deregister(info *registry.Info) error {
	if err := n.validRegistryInfo(info); err != nil {
		return fmt.Errorf("valid parse registry info error: %w", err)
	}
	host, port, err := net.SplitHostPort(info.Addr.String())
	if err != nil {
		return err
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("parse registry info port error: %w", err)
	}
	ip := net.ParseIP(host)
	if ip == nil || ip.IsUnspecified() {
		host = utils.LocalIP()
	}
	success, err := n.client.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          host,
		Port:        uint64(portInt),
		ServiceName: info.ServiceName,
		GroupName:   n.opts.group,
		Cluster:     n.opts.cluster,
		Ephemeral:   true,
	})
	if success {
		hlog.SystemLogger().Info("deregister instance success")
	}
	if err != nil {
		return err
	}
	return nil
}

func (n *nacosRegistry) validRegistryInfo(info *registry.Info) error {
	if info == nil {
		return fmt.Errorf("*registry.Info can not be empty")
	}
	if info.ServiceName == "" {
		return fmt.Errorf("*registry.Info ServiceName can not be empty")
	}
	if info.Addr == nil {
		return fmt.Errorf("*registry.Info GetAddr can not be empty")
	}
	return nil
}
