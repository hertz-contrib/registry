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
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/go-zookeeper/zk"
	"github.com/hertz-contrib/registry/zookeeper/utils"
)

const (
	Separator = "/"
	Scheme    = "digest" // For auth
)

type RegistryEntity struct {
	Weight int
	Tags   map[string]string
}

type zookeeperRegistry struct {
	conn           *zk.Conn
	authOpen       bool
	user, password string
}

func (z *zookeeperRegistry) Register(info *registry.Info) error {
	if err := z.validRegistryInfo(info); err != nil {
		return fmt.Errorf("valid parse registry info error: %w", err)
	}
	path, err := buildPath(info)
	if err != nil {
		return err
	}
	content, err := json.Marshal(&RegistryEntity{Weight: info.Weight, Tags: info.Tags})
	if err != nil {
		return err
	}
	return z.createNode(path, content, true)
}

func (z *zookeeperRegistry) Deregister(info *registry.Info) error {
	if err := z.validRegistryInfo(info); err != nil {
		return fmt.Errorf("valid parse registry info error: %w", err)
	}

	path, err := buildPath(info)
	if err != nil {
		return err
	}
	return z.deleteNode(path)
}

func NewZookeeperRegistry(servers []string, sessionTimeout time.Duration) (registry.Registry, error) {
	conn, _, err := zk.Connect(servers, sessionTimeout)
	if err != nil {
		return nil, err
	}
	return &zookeeperRegistry{conn: conn}, nil
}

func NewZookeeperRegistryWithAuth(servers []string, sessionTimeout time.Duration, user, password string) (registry.Registry, error) {
	if user == "" || password == "" {
		return nil, fmt.Errorf("user or password can't be empty")
	}
	conn, _, err := zk.Connect(servers, sessionTimeout)
	if err != nil {
		return nil, err
	}
	auth := []byte(fmt.Sprintf("%s:%s", user, password))
	err = conn.AddAuth(Scheme, auth)
	if err != nil {
		return nil, err
	}
	return &zookeeperRegistry{conn: conn, authOpen: true, user: user, password: password}, nil
}

func (z *zookeeperRegistry) validRegistryInfo(info *registry.Info) error {
	if info == nil {
		return errors.New("registry.Info can not be empty")
	}
	if info.ServiceName == "" {
		return errors.New("registry.Info ServiceName can not be empty")
	}
	if info.Addr == nil {
		return errors.New("registry.Info Addr can not be empty")
	}
	return nil
}

// buildPath path format as follows: {serviceName}/{ip}:{port}
func buildPath(info *registry.Info) (string, error) {
	var path string
	if !strings.HasPrefix(info.ServiceName, Separator) {
		path = Separator + info.ServiceName
	}

	if host, port, err := net.SplitHostPort(info.Addr.String()); err == nil {
		if port == "" {
			return "", fmt.Errorf("registry info addr missing port")
		}
		if host == "" {
			ipv4, err := utils.GetLocalIPv4Address()
			if err != nil {
				return "", fmt.Errorf("get local ipv4 error, cause %w", err)
			}
			path = path + Separator + ipv4 + ":" + port
		} else {
			path = path + Separator + host + ":" + port
		}
	} else {
		return "", fmt.Errorf("parse registry info addr error")
	}
	return path, nil
}

func (z *zookeeperRegistry) createNode(path string, content []byte, ephemeral bool) error {
	i := strings.LastIndex(path, Separator)
	if i > 0 {
		err := z.createNode(path[0:i], nil, false)
		if err != nil && !errors.Is(err, zk.ErrNodeExists) {
			return err
		}
	}
	var flag int32
	if ephemeral {
		flag = zk.FlagEphemeral
	}
	var acl []zk.ACL
	if z.authOpen {
		acl = zk.DigestACL(zk.PermAll, z.user, z.password)
	} else {
		acl = zk.WorldACL(zk.PermAll)
	}
	_, err := z.conn.Create(path, content, flag, acl)
	if err != nil {
		return fmt.Errorf("create node [%s] error, cause %w", path, err)
	}
	return nil
}

func (z *zookeeperRegistry) deleteNode(path string) error {
	err := z.conn.Delete(path, -1)
	if err != nil && err != zk.ErrNoNode {
		return fmt.Errorf("delete node [%s] error, cause %w", path, err)
	}
	return nil
}
