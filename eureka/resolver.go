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

package eureka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/hudl/fargo"
)

var _ discovery.Resolver = (*eurekaResolver)(nil)

// eurekaResolver is a resolver using eureka.
type eurekaResolver struct {
	eurekaConn *fargo.EurekaConnection
}

// NewEurekaResolver creates a eureka resolver with a slice of server addresses.
func NewEurekaResolver(servers []string) *eurekaResolver {
	conn := fargo.NewConn(servers...)

	return &eurekaResolver{eurekaConn: &conn}
}

// NewEurekaResolverFromConfig creates a eureka resolver with given configuration.
func NewEurekaResolverFromConfig(config fargo.Config) *eurekaResolver {
	conn := fargo.NewConnFromConfig(config)

	return &eurekaResolver{
		eurekaConn: &conn,
	}
}

// NewEurekaResolverFromConn creates a eureka resolver using an existing connection.
func NewEurekaResolverFromConn(conn fargo.EurekaConnection) *eurekaResolver {
	return &eurekaResolver{
		eurekaConn: &conn,
	}
}

// Target implements the Resolver interface.
func (r *eurekaResolver) Target(ctx context.Context, target *discovery.TargetInfo) string {
	return target.Host
}

// Resolve implements the Resolver interface.
func (r *eurekaResolver) Resolve(ctx context.Context, desc string) (discovery.Result, error) {
	application, err := r.eurekaConn.GetApp(desc)
	if err != nil {
		if errors.As(err, &fargo.AppNotFoundError{}) {
			return discovery.Result{}, fmt.Errorf("app not found [%s]", desc)
		}
		return discovery.Result{}, err
	}

	eurekaInstances := application.Instances
	instances, err := r.getInstances(eurekaInstances)
	if err != nil {
		return discovery.Result{}, err
	}

	return discovery.Result{CacheKey: desc, Instances: instances}, nil
}

// Name implements the Resolver interface.
func (r *eurekaResolver) Name() string {
	return Eureka
}

func (r *eurekaResolver) getInstances(instances []*fargo.Instance) ([]discovery.Instance, error) {
	res := make([]discovery.Instance, 0, len(instances))
	for _, instance := range instances {
		dInstance, err := r.getInstance(instance)
		if err != nil {
			return nil, err
		}
		res = append(res, dInstance)
	}

	return res, nil
}

func (r *eurekaResolver) getInstance(instance *fargo.Instance) (discovery.Instance, error) {
	var dInstance discovery.Instance
	var e RegistryEntity
	meta, err := instance.Metadata.GetString(Meta)
	if err != nil {
		return dInstance, err
	}
	if err = json.Unmarshal([]byte(meta), &e); err != nil {
		return dInstance, err
	}

	return discovery.NewInstance(TCP, fmt.Sprintf("%s:%d", instance.IPAddr, instance.Port), e.Weight, e.Tags), nil
}
