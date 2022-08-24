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

package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"go.etcd.io/etcd/clientv3"
)

var _ registry.Registry = (*etcdRegistry)(nil)

const (
	Separator = "/"
)

type etcdRegistry struct {
	client  *clientv3.Client
	timeout time.Duration
}

type RegistryEntity struct {
	Weight int
	Tags   map[string]string
}

func (e etcdRegistry) Register(info *registry.Info) error {
	path, err := buildPath(info)
	if err != nil {
		return err
	}
	content, err := json.Marshal(&RegistryEntity{Weight: info.Weight, Tags: info.Tags})
	if err != nil {
		return err
	}
	return e.addNode(path, string(content))
}

func (e etcdRegistry) Deregister(info *registry.Info) error {
	if info == nil {
		return fmt.Errorf("registry info can't be nil")
	}
	path, err := buildPath(info)
	if err != nil {
		return err
	}
	return e.delNode(path)
}

// buildPath path format as follows: {serviceName}/{ip}:{port}
func buildPath(info *registry.Info) (string, error) {
	var path string
	if !strings.HasPrefix(info.ServiceName, Separator) {
		path = Separator + info.ServiceName
	}
	host, port, err := net.SplitHostPort(info.Addr.String())
	if err != nil {
		return "", fmt.Errorf("parse registry info addr error")
	}
	if port == "" {
		return "", fmt.Errorf("registry info addr missing port")
	}
	if host == "" || host == "::" {
		host = utils.LocalIP()
	}
	path = path + Separator + net.JoinHostPort(host, port)

	return path, nil
}

func (e *etcdRegistry) addNode(path string, content string) error {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	_, err := e.client.Put(ctx, path, content)
	cancel()
	if err != nil {
		return fmt.Errorf("create node [%s] error, cause %w", path, err)
	}
	return nil
}

func (e *etcdRegistry) delNode(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	_, err := e.client.Delete(ctx, path)
	cancel()
	if err != nil {
		return fmt.Errorf("delete node [%s] error, cause %w", path, err)
	}
	return nil
}

func NewEtcdRegistry(cli *clientv3.Client, sessionTimeout time.Duration) (registry.Registry, error) {
	return &etcdRegistry{cli, sessionTimeout}, nil
}
