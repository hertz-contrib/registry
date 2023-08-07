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

package common

import (
	"os"
	"strconv"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

const (
	nacosEnvServerAddr     = "serverAddr"
	nacosEnvServerPort     = "serverPort"
	nacosEnvNamespaceID    = "namespace"
	nacosDefaultServerAddr = "127.0.0.1"
	nacosDefaultPort       = 8848
	nacosDefaultRegionID   = "cn-hangzhou"
)

// NewDefaultNacosConfig create a basic Nacos client.
func NewDefaultNacosConfig() (naming_client.INamingClient, error) {
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(NacosAddr(), uint64(NacosPort())),
	}
	cc := constant.ClientConfig{
		NamespaceId:         NacosNameSpaceID(),
		RegionId:            nacosDefaultRegionID,
		CustomLogger:        NewCustomNacosLogger(),
		NotLoadCacheAtStart: true,
	}
	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NacosPort Get Nacos port from environment variables.
func NacosPort() int64 {
	portText := os.Getenv(nacosEnvServerPort)
	if len(portText) == 0 {
		return nacosDefaultPort
	}
	port, err := strconv.ParseInt(portText, 10, 64)
	if err != nil {
		hlog.Errorf("ParseInt failed,err:%s", err.Error())
		return nacosDefaultPort
	}
	return port
}

// NacosAddr Get Nacos addr from environment variables.
func NacosAddr() string {
	addr := os.Getenv(nacosEnvServerAddr)
	if len(addr) == 0 {
		return nacosDefaultServerAddr
	}
	return addr
}

// NacosNameSpaceID Get Nacos namespace id from environment variables.
func NacosNameSpaceID() string {
	return os.Getenv(nacosEnvNamespaceID)
}
