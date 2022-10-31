package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

var gm *mentor

var form = make(map[string]addrs)

type addrs []string

type mentor struct {
	mform map[string]addrs
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
	sub := r.client.Subscribe(ctx, fmt.Sprintf("/%s/%s/%s", hertz, info.ServiceName, server))
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
				m.insertForm(split[1], split[2])
				hlog.Infof("HERTZ: service info %v", m.mform)
			} else if split[0] == deregister {
				m.removeAddr(split[1], split[2])
				hlog.Infof("HERTZ: service info %v", m.mform)
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
				m.removeService(info.ServiceName)
			}
		case <-ctx.Done():
			break
		default:
		}
	}
}

func (m *mentor) insertForm(serviceName string, addr string) {
	m.mform[serviceName] = append(m.mform[serviceName], addr)
}

func (m *mentor) removeService(serviceName string) {
	delete(m.mform, serviceName)
}

func (m *mentor) removeAddr(serviceName string, addr string) {
	for i, v := range m.mform[serviceName] {
		if v == addr {
			m.mform[serviceName] = append(m.mform[serviceName][:i], m.mform[serviceName][i+1:]...)
		}
	}
}
