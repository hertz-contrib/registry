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

package redis

import (
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
)

const (
	hertz  = "hertz"
	server = "server"
	tcp    = "tcp"
)

type registryHash struct {
	key   string
	field string
	value string
}

type registryInfo struct {
	ServiceName string            `json:"service_name"`
	Addr        string            `json:"addr"`
	Weight      int               `json:"weight"`
	Tags        map[string]string `json:"tags"`
}

func generateKey(serviceName, serviceType string) string {
	return fmt.Sprintf("/%s/%s/%s", hertz, serviceName, serviceType)
}

func prepareRegistryHash(info *registry.Info) (*registryHash, error) {
	meta, err := sonic.Marshal(convertInfo(info))
	if err != nil {
		return nil, err
	}
	return &registryHash{
		key:   generateKey(info.ServiceName, server),
		field: info.Addr.String(),
		value: string(meta),
	}, nil
}

func convertInfo(info *registry.Info) *registryInfo {
	return &registryInfo{
		ServiceName: info.ServiceName,
		Addr:        info.Addr.String(),
		Weight:      info.Weight,
		Tags:        info.Tags,
	}
}
