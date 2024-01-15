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
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
)

const kvJoinChar = ":"

var errIllegalTagChar = errors.New("illegal tag character")

func parseAddr(addr net.Addr) (host string, port int, err error) {
	host, portStr, err := net.SplitHostPort(addr.String())
	if err != nil {
		return "", 0, fmt.Errorf("calling net.SplitHostPort failed, addr: %s, err: %w", addr.String(), err)
	}

	if host == "" || host == "::" {
		detectHost := utils.LocalIP()
		if detectHost == utils.UNKNOWN_IP_ADDR {
			return "", 0, fmt.Errorf("get local ip error")
		}

		host, _, err = net.SplitHostPort(detectHost)

		if err != nil {
			return "", 0, fmt.Errorf("empty host")
		}
	}

	port, err = strconv.Atoi(portStr)
	if err != nil {
		return "", 0, fmt.Errorf("parsing registry info port failed, portStr:%s, err: %w", portStr, err)
	}
	if port == 0 {
		return "", 0, fmt.Errorf("invalid port %d", port)
	}

	return host, port, nil
}

func getServiceId(info *registry.Info) (string, error) {
	host, port, err := parseAddr(info.Addr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s:%d", info.ServiceName, host, port), nil
}

// convTagMapToSlice Tags map be converted to slice.
// Keys must not contain `:`.
func convTagMapToSlice(tagMap map[string]string) ([]string, error) {
	svcTags := make([]string, 0, len(tagMap))
	for k, v := range tagMap {
		var tag string
		if strings.Contains(k, kvJoinChar) {
			return svcTags, errIllegalTagChar
		}
		if v == "" {
			tag = k
		} else {
			tag = fmt.Sprintf("%s%s%s", k, kvJoinChar, v)
		}
		svcTags = append(svcTags, tag)
	}
	return svcTags, nil
}

// splitTags Tags characters be separated to map.
func splitTags(tags []string) map[string]string {
	n := len(tags)
	tagMap := make(map[string]string, n)
	if n == 0 {
		return tagMap
	}

	for _, tag := range tags {
		if tag == "" {
			continue
		}
		strArr := strings.SplitN(tag, kvJoinChar, 2)
		if len(strArr) == 2 {
			key := strArr[0]
			tagMap[key] = strArr[1]
		}
		if len(strArr) == 1 {
			tagMap[strArr[0]] = ""
		}
	}

	return tagMap
}
