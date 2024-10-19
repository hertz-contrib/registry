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
	"net"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/stretchr/testify/assert"
)

// TestEurekaRegistryAndDeRegistry register one or more instances for each service,
// check if result of service discovery matches what have been registered.
// Then tearing down instance one by one, validate the number of available instances during the process.
func TestEurekaRegistryAndDeRegistry(t *testing.T) {
	tests := []struct {
		info    []*registry.Info
		target  discovery.TargetInfo
		wantErr bool
	}{
		{
			// register single instance
			info: []*registry.Info{
				{
					ServiceName: "hertz.discovery.single",
					Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8890},
					Weight:      10,
					Tags:        map[string]string{"region": "eu-south-1"},
				},
			},
			wantErr: false,
			target: discovery.TargetInfo{
				Host: "hertz.discovery.single",
				Tags: nil,
			},
		},
		{
			// register multiple instances
			info: []*registry.Info{
				{
					ServiceName: "hertz.discovery.multiple",
					Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8890},
					Weight:      15,
					Tags:        nil,
				},
				{
					ServiceName: "hertz.discovery.multiple",
					Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8891},
					Weight:      20,
					Tags:        nil,
				},
			},
			wantErr: false,
			target: discovery.TargetInfo{
				Host: "hertz.discovery.multiple",
				Tags: nil,
			},
		},
	}

	for _, tes := range tests {
		r := NewEurekaRegistry([]string{"http://127.0.0.1:8761/eureka"}, 11*time.Second)
		addrMap := map[string]*registry.Info{}

		for _, info := range tes.info {

			var err error
			addrMap[info.Addr.String()] = info
			if err := r.Register(info); err != nil {
				t.Errorf("info register err")
			}
			assert.False(t, err != nil)

		}

		resolver := NewEurekaResolver([]string{"http://127.0.0.1:8761/eureka"})
		result, err := resolver.Resolve(context.Background(), tes.target.Host)
		assert.Nil(t, err)
		assert.Equal(t, len(tes.info), len(result.Instances))

		// iterate over result to check metadata of each instance
		for _, instance := range result.Instances {
			addr := instance.Address().String()
			info, ok := addrMap[addr]

			assert.Equal(t, true, ok)
			assert.Equal(t, info.Addr.String(), addr)
			assert.Equal(t, info.Weight, instance.Weight())
			assert.Equal(t, info.ServiceName, tes.target.Host)

			// check all tags have been preserved
			for k, expected := range info.Tags {
				actual, exist := instance.Tag(k)
				assert.Equal(t, true, exist)
				assert.Equal(t, expected, actual)
			}
		}

		// keep track the number of instance removed from eureka
		var instanceRemoved int

		for _, info := range tes.info {

			if err := r.Deregister(info); err != nil {
				t.Errorf("info deregister err")
			}
			assert.Nil(t, err)
			instanceRemoved++

			result, err := resolver.Resolve(context.Background(), tes.target.Host)

			// if all instance have been removed, returns app not found error
			if instanceRemoved == len(tes.info) {
				assert.True(t, err != nil)
			}
			assert.Equal(t, len(result.Instances), len(tes.info)-instanceRemoved)

		}

	}
}

// TestEurekaRegisterWithDefaultWeight test if the default weight has been assigned to instance.
func TestEurekaRegisterWithDefaultWeight(t *testing.T) {
	info := &registry.Info{
		ServiceName: "hertz.discovery.default_weight",
		Addr:        &net.TCPAddr{Port: 8890},
		Tags:        nil,
	}

	target := discovery.TargetInfo{
		Host: "hertz.discovery.default_weight",
		Tags: nil,
	}

	r := NewEurekaRegistry([]string{"http://127.0.0.1:8761/eureka"}, 11*time.Second)

	var err error
	if err := r.Register(info); err != nil {
		t.Errorf("info register err")
	}
	assert.Nil(t, err)

	resolver := NewEurekaResolver([]string{"http://127.0.0.1:8761/eureka"})
	result, err := resolver.Resolve(context.Background(), target.Host)
	assert.Nil(t, err)
	assert.Equal(t, len(result.Instances), 1)
	assert.Equal(t, result.Instances[0].Weight(), registry.DefaultWeight)
}

// TestEurekaRegistryWithInvalidInstanceInfo run Register against a collection of invalid instance
// in these cases, instance registration should fail.
func TestEurekaRegistryWithInvalidInstanceInfo(t *testing.T) {
	tests := []struct {
		info        *registry.Info
		expectedErr error
	}{
		{
			// invalid service name
			info: &registry.Info{
				ServiceName: "",
				Weight:      10,
				Tags:        map[string]string{"idc": "hl"},
				Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8888},
			},
			expectedErr: ErrEmptyServiceName,
		},
		{
			// service info is nil
			info:        nil,
			expectedErr: ErrNilInfo,
		},
		{
			// address is nil
			info: &registry.Info{
				ServiceName: "test",
				Weight:      10,
				Tags:        map[string]string{"idc": "hl"},
				Addr:        nil,
			},
			expectedErr: ErrNilAddr,
		},
		{
			// port is missing
			info: &registry.Info{
				ServiceName: "test",
				Weight:      10,
				Tags:        map[string]string{"idc": "hl"},
				Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)},
			},
			expectedErr: ErrMissingPort,
		},
	}

	r := NewEurekaRegistry([]string{"http://127.0.0.1:8761/eureka"}, 11*time.Second)

	for _, entry := range tests {
		registerErr := r.Register(entry.info)
		assert.True(t, registerErr != nil)
		assert.Equal(t, entry.expectedErr, registerErr)
	}
}
