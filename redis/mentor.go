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
	"context"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

var gm *mentor

var form = make(map[string]addrs)

type addrs []string

type mentor struct {
	mform map[string]addrs
	mu    sync.Mutex
}

// newMentor use singleton
func newMentor() *mentor {
	if gm != nil {
		return gm
	}
	m := &mentor{mform: form}
	gm = m
	return gm
}

func (m *mentor) subscribe(ctx context.Context, info *registry.Info, r *redisRegistry) {
	sub := r.client.Subscribe(ctx, generateKey(info.ServiceName, server))
	defer sub.Close()
	r.wg.Done()
	select {
	case <-ctx.Done():
		return
	default:
		ch := sub.Channel()
		for msg := range ch {
			split := strings.Split(msg.Payload, "-")
			if split[0] == register {
				m.mu.Lock()
				m.insertForm(split[1], split[2])
				hlog.Infof("HERTZ: service info %v", m.mform)
				m.mu.Unlock()
			} else if split[0] == deregister {
				m.mu.Lock()
				m.removeAddr(split[1], split[2])
				hlog.Infof("HERTZ: service info %v", m.mform)
				m.mu.Unlock()
			} else {
				hlog.Warnf("HERTZ: invalid message %v", msg)
			}
		}
	}
}

func (m *mentor) monitorTTL(ctx context.Context, hash *registryHash, info *registry.Info, r *redisRegistry) {
	ticker := time.NewTicker(defaultMonitorTime)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if r.client.TTL(ctx, hash.key).Val() == -2 {
				m.mu.Lock()
				m.removeService(info.ServiceName)
				m.mu.Unlock()
			}
		case <-ctx.Done():
			break
		}
	}
}

func (m *mentor) insertForm(serviceName, addr string) {
	m.mform[serviceName] = append(m.mform[serviceName], addr)
}

func (m *mentor) removeService(serviceName string) {
	delete(m.mform, serviceName)
}

func (m *mentor) removeAddr(serviceName, addr string) {
	for i, v := range m.mform[serviceName] {
		if v == addr {
			m.mform[serviceName] = append(m.mform[serviceName][:i], m.mform[serviceName][i+1:]...)
		}
	}
}
