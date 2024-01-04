// Copyright 2023 CloudWeGo Authors.
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
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

const (
	nacosV2EnvServerAddr     = "serverAddr"
	nacosV2EnvServerPort     = "serverPort"
	nacosV2EnvNamespaceID    = "namespace"
	nacosV2DefaultServerAddr = "127.0.0.1"
	nacosV2DefaultPort       = 8848
	nacosV2DefaultRegionID   = "cn-hangzhou"
)

// NewDefaultNacosV2Config create a default Nacos client. & server.
func NewDefaultNacosV2Config() (naming_client.INamingClient, error) {
	clientConfig := constant.ClientConfig{
		NamespaceId:         NacosV2NameSpaceID(),
		RegionId:            nacosV2DefaultRegionID,
		NotLoadCacheAtStart: true,
	}
	serverConfig := []constant.ServerConfig{
		*constant.NewServerConfig(NacosV2Addr(), uint64(NacosV2Port())),
	}
	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfig,
		},
	)
	if err != nil {
		return nil, err
	}
	return namingClient, nil
}

// NacosV2NameSpaceID Get Nacos namespace id from environment variables.
func NacosV2NameSpaceID() string {
	return os.Getenv(nacosV2EnvNamespaceID)
}

// NacosV2Addr Get Nacos addr from environment variables.
func NacosV2Addr() string {
	addr := os.Getenv(nacosV2EnvServerAddr)
	if len(addr) == 0 {
		return nacosV2DefaultServerAddr
	}
	return addr
}

// NacosV2Port Get Nacos port from environment variables.
func NacosV2Port() int64 {
	portText := os.Getenv(nacosV2EnvServerPort)
	if len(portText) == 0 {
		return nacosV2DefaultPort
	}
	port, err := strconv.ParseInt(portText, 10, 64)
	if err != nil {
		hlog.Errorf("ParseInt failed,err:%s", err.Error())
		return nacosV2DefaultPort
	}
	return port
}
