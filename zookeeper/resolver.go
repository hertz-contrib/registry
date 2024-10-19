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
	"github.com/cloudwego/hertz/pkg/app/client/discovery"
)

// NewZookeeperResolver create a zookeeper based resolver
func NewZookeeperResolver(servers []string, sessionTimeout time.Duration) (discovery.Resolver, error) {
	return zookeeperhertz.NewZookeeperResolver(servers, sessionTimeout)
}

// NewZookeeperResolver create a zookeeper based resolver with auth
func NewZookeeperResolverWithAuth(servers []string, sessionTimeout time.Duration, user, password string) (discovery.Resolver, error) {
	return zookeeperhertz.NewZookeeperResolverWithAuth(servers, sessionTimeout, user, password)
}
