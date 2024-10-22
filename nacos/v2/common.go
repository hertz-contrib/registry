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
	"os"
	"strconv"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

const (
	nacosEnvServerAddr     = "serverAddr"
	nacosEnvServerPort     = "serverPort"
	nacosEnvNamespaceID    = "namespace"
	nacosDefaultServerAddr = "127.0.0.1"
	nacosDefaultPort       = 8848
	nacosDefaultRegionID   = "cn-hangzhou"
)

// GetPort Get Nacos port from environment variables.
func GetPort() int64 {
	portText := os.Getenv(nacosEnvServerPort)
	if len(portText) == 0 {
		return nacosDefaultPort
	}
	port, err := strconv.ParseInt(portText, 10, 64)
	if err != nil {
		hlog.Errorf("ParseInt failed, err:%s", err.Error())
		return nacosDefaultPort
	}
	return port
}

// GetAddr Get Nacos addr from environment variables.
func GetAddr() string {
	addr := os.Getenv(nacosEnvServerAddr)
	if len(addr) == 0 {
		return nacosDefaultServerAddr
	}
	return addr
}

// GetNameSpaceID Get Nacos namespace id from environment variables.
func GetNameSpaceID() string {
	return os.Getenv(nacosEnvNamespaceID)
}

// compareMaps compares two maps regardless of nil or empty
func compareMaps(m1, m2 map[string]string) bool {
	// if both maps are nil, they are equal
	if m1 == nil && m2 == nil {
		return true
	}
	// if the lengths are different, the maps are not equal
	if len(m1) != len(m2) {
		return false
	}
	// iterate over the keys of m1 and check if they exist in m2 with the same value
	for k, v := range m1 {
		if v2, ok := m2[k]; !ok || v != v2 {
			return false
		}
	}
	// return true if no differences are found
	return true
}
