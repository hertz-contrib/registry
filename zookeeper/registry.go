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

package zookeeper

import (
	"time"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/zookeeper/zookeeperhertz"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
)

const (
	Separator = "/"
	Scheme    = "digest" // For auth
)

type RegistryEntity struct {
	Weight int
	Tags   map[string]string
}

func NewZookeeperRegistry(servers []string, sessionTimeout time.Duration) (registry.Registry, error) {
	return zookeeperhertz.NewZookeeperRegistry(servers, sessionTimeout)
}

func NewZookeeperRegistryWithAuth(servers []string, sessionTimeout time.Duration, user, password string) (registry.Registry, error) {
	return zookeeperhertz.NewZookeeperRegistryWithAuth(servers, sessionTimeout, user, password)
}
