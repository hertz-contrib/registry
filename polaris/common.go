/*
 * Copyright 2021 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package polaris

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/config"
	"github.com/polarismesh/polaris-go/pkg/model"
)

// GetPolarisConfig get polaris config from endpoints.
func GetPolarisConfig(configFile ...string) (api.SDKContext, error) {
	var (
		cfg config.Configuration
		err error
	)

	if len(configFile) != 0 {
		cfg, err = config.LoadConfigurationByFile(configFile[0])
	} else {
		cfg, err = config.LoadConfigurationByDefaultFile()
	}

	if err != nil {
		return nil, err
	}

	sdkCtx, err := api.InitContextByConfig(cfg)
	if err != nil {
		return nil, err
	}

	return sdkCtx, nil
}

// SplitDescription splits description to namespace and serviceName.
func SplitDescription(description string) (string, string) {
	str := strings.Split(description, ":")
	namespace, serviceName := str[0], str[1]
	return namespace, serviceName
}

// ChangePolarisInstanceToHertz transforms polaris instance to Hertz instance.
func ChangePolarisInstanceToHertz(PolarisInstance model.Instance) discovery.Instance {
	weight := PolarisInstance.GetWeight()
	if weight <= 0 {
		weight = registry.DefaultWeight
	}
	addr := net.JoinHostPort(PolarisInstance.GetHost(), strconv.Itoa(int(PolarisInstance.GetPort())))

	tags := map[string]string{
		"namespace": PolarisInstance.GetNamespace(),
	}

	HertzInstance := discovery.NewInstance(PolarisInstance.GetProtocol(), addr, weight, tags)
	// In HertzInstance, tags can be used as IDC、Cluster、Env、namespace、and so on.
	return HertzInstance
}

// GetInfoHostAndPort gets Host and port from info.Addr.
func GetInfoHostAndPort(Addr string) (string, int, error) {
	infoHost, port, err := net.SplitHostPort(Addr)
	if err != nil {
		return "", 0, err
	}
	if port == "" {
		return infoHost, 0, fmt.Errorf("registry info addr missing port")
	}
	if infoHost == "" || infoHost == "::" {
		infoHost = utils.LocalIP()
		if infoHost == utils.UNKNOWN_IP_ADDR {
			return "", 0, fmt.Errorf("get local ip error")
		}
	}
	infoPort, err := strconv.Atoi(port)
	if err != nil {
		return "", 0, err
	}
	return infoHost, infoPort, nil
}

// GetInstanceKey generates instanceKey for one instance.
func GetInstanceKey(namespace, serviceName, host, port string) string {
	var instanceKey strings.Builder
	instanceKey.WriteString(namespace)
	instanceKey.WriteString(":")
	instanceKey.WriteString(serviceName)
	instanceKey.WriteString(":")
	instanceKey.WriteString(host)
	instanceKey.WriteString(":")
	instanceKey.WriteString(port)
	return instanceKey.String()
}
