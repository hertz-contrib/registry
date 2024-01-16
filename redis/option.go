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

	"github.com/redis/go-redis/v9"
)

type Option func(opts *Options)

type Options struct {
	*redis.Options
	expireTime      int
	refreshInterval int
}

// WithExpireTime redis key expiration time in seconds
// NOTE: expiration time must be greater than refresh interval
// Default: 60s
func WithExpireTime(time int) Option {
	return func(opts *Options) {
		opts.expireTime = time
	}
}

// WithRefreshInterval redis key refresh interval in seconds
// NOTE: refresh interval must be less than expiration time
// Default: 30s
func WithRefreshInterval(interval int) Option {
	return func(opts *Options) {
		opts.refreshInterval = interval
	}
}

func WithPassword(password string) Option {
	return func(opts *Options) {
		opts.Password = password
	}
}

func WithDB(db int) Option {
	return func(opts *Options) {
		opts.DB = db
	}
}

func WithTLSConfig(t *tls.Config) Option {
	return func(opts *Options) {
		opts.TLSConfig = t
	}
}

func WithDialer(dialer func(ctx context.Context, network, addr string) (net.Conn, error)) Option {
	return func(opts *Options) {
		opts.Dialer = dialer
	}
}

func WithReadTimeout(t time.Duration) Option {
	return func(opts *Options) {
		opts.ReadTimeout = t
	}
}

func WithWriteTimeout(t time.Duration) Option {
	return func(opts *Options) {
		opts.WriteTimeout = t
	}
}
