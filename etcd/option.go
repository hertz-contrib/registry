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
	"io/ioutil"
	"time"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/etcd/etcdhertz"
)

const (
	defaultTTL = 60
)

type option struct {
	// etcd client config
	options []etcdhertz.Option
}

/*type retryCfg struct {
	// The maximum number of call attempt times, including the initial call
	maxAttemptTimes uint
	// observeDelay is the delay time for checking the service status under normal conditions
	observeDelay time.Duration
	// retryDelay is the delay time for attempting to register the service after disconnecting
	retryDelay time.Duration
}*/

type Option = etcdhertz.Option

// WithMaxAttemptTimes sets the maximum number of call attempt times, including the initial call
func WithMaxAttemptTimes(maxAttemptTimes uint) Option {
	return etcdhertz.WithMaxAttemptTimes(maxAttemptTimes)
}

// WithObserveDelay sets the delay time for checking the service status under normal conditions
func WithObserveDelay(observeDelay time.Duration) Option {
	return etcdhertz.WithObserveDelay(observeDelay)
}

// WithRetryDelay sets the delay time of retry
func WithRetryDelay(t time.Duration) Option {
	return etcdhertz.WithRetryDelay(t)
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

// WithTLSOpt returns a option that authentication by tls/ssl.
func WithTLSOpt(certFile, keyFile, caFile string) Option {
	return etcdhertz.WithTLSOpt(certFile, keyFile, caFile)
}

// WithAuthOpt returns an option that authentication by username and password.
func WithAuthOpt(username, password string) Option {
	return etcdhertz.WithAuthOpt(username, password)
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
