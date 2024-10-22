// Copyright 2023 CloudWeGo Authors.
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
	"crypto/tls"
	"net"
	"time"

	redishertz "github.com/cloudwego-contrib/cwgo-pkg/registry/redis/redishertz"
)

type Option = redishertz.Option

// WithExpireTime redis key expiration time in seconds
// NOTE: expiration time must be greater than refresh interval
// Default: 60s
func WithExpireTime(time int) Option {
	return redishertz.WithExpireTime(time)
}

// WithRefreshInterval redis key refresh interval in seconds
// NOTE: refresh interval must be less than expiration time
// Default: 30s
func WithRefreshInterval(interval int) Option {
	return redishertz.WithRefreshInterval(interval)
}

func WithPassword(password string) Option {
	return redishertz.WithPassword(password)
}

func WithDB(db int) Option {
	return redishertz.WithDB(db)
}

func WithTLSConfig(t *tls.Config) Option {
	return redishertz.WithTLSConfig(t)
}

func WithDialer(dialer func(ctx context.Context, network, addr string) (net.Conn, error)) Option {
	return redishertz.WithDialer(dialer)
}

func WithReadTimeout(t time.Duration) Option {
	return redishertz.WithReadTimeout(t)
}

func WithWriteTimeout(t time.Duration) Option {
	return redishertz.WithWriteTimeout(t)
}
