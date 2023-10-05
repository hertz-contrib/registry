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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	defaultTTL = 60
)

type option struct {
	// etcd client config
	etcdCfg  clientv3.Config
	retryCfg *retryCfg
}

type retryCfg struct {
	// The maximum number of call attempt times, including the initial call
	maxAttemptTimes uint
	// observeDelay is the delay time for checking the service status under normal conditions
	observeDelay time.Duration
	// retryDelay is the delay time for attempting to register the service after disconnecting
	retryDelay time.Duration
}

type Option func(o *option)

// WithMaxAttemptTimes sets the maximum number of call attempt times, including the initial call
func WithMaxAttemptTimes(maxAttemptTimes uint) Option {
	return func(o *option) {
		o.retryCfg.maxAttemptTimes = maxAttemptTimes
	}
}

// WithObserveDelay sets the delay time for checking the service status under normal conditions
func WithObserveDelay(observeDelay time.Duration) Option {
	return func(o *option) {
		o.retryCfg.observeDelay = observeDelay
	}
}

// WithRetryDelay sets the delay time of retry
func WithRetryDelay(t time.Duration) Option {
	return func(o *option) {
		o.retryCfg.retryDelay = t
	}
}

func (o *option) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

// instanceInfo used to stored service basic info in etcd.
type instanceInfo struct {
	Network string            `json:"network"`
	Address string            `json:"address"`
	Weight  int               `json:"weight"`
	Tags    map[string]string `json:"tags"`
}

func serviceKeyPrefix(serviceName string) string {
	return etcdPrefix + "/" + serviceName
}

// serviceKey generates the key used to stored in etcd.
func serviceKey(serviceName, addr string) string {
	return serviceKeyPrefix(serviceName) + "/" + addr
}

// validateRegistryInfo validate the registry.Info
func validateRegistryInfo(info *registry.Info) error {
	if info == nil {
		return fmt.Errorf("registry.Info can not be empty")
	}
	if info.ServiceName == "" {
		return fmt.Errorf("registry.Info ServiceName can not be empty")
	}
	if info.Addr == nil {
		return fmt.Errorf("registry.Info Addr can not be empty")
	}
	return nil
}

// getTTL get the lease from default or from the env
func getTTL() int64 {
	var ttl int64 = defaultTTL
	if str, ok := os.LookupEnv(ttlKey); ok {
		if t, err := strconv.Atoi(str); err == nil {
			ttl = int64(t)
		}
	}
	return ttl
}

// WithTLSOpt returns a option that authentication by tls/ssl.
func WithTLSOpt(certFile, keyFile, caFile string) Option {
	return func(o *option) {
		tlsCfg, err := newTLSConfig(certFile, keyFile, caFile, "")
		if err != nil {
			hlog.Errorf("HERTZ: tls failed with err: %v , skipping tls.", err)
		}
		o.etcdCfg.TLS = tlsCfg
	}
}

// WithAuthOpt returns an option that authentication by username and password.
func WithAuthOpt(username, password string) Option {
	return func(o *option) {
		o.etcdCfg.Username = username
		o.etcdCfg.Password = password
	}
}

func newTLSConfig(certFile, keyFile, caFile, serverName string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	successful := caCertPool.AppendCertsFromPEM(caCert)
	if !successful {
		return nil, errors.New("failed to parse ca certificate as PEM encoded content")
	}
	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
	return cfg, nil
}
