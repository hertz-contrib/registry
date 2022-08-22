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

package consul

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

var (
	consulClient *consulapi.Client
	cRegistry    registry.Registry
	cResolver    discovery.Resolver
	consulAddr   = "127.0.0.1:8500"
	localIpAddr  = "127.0.0.1"
)

func init() {
	config1 := consulapi.DefaultConfig()
	config1.Address = consulAddr
	cli1, err := consulapi.NewClient(config1)
	if err != nil {
		log.Fatal(err)
		return
	}
	cRegistry = NewConsulRegister(cli1)

	config2 := consulapi.DefaultConfig()
	config2.Address = consulAddr
	cli2, err := consulapi.NewClient(config2)
	if err != nil {
		log.Fatal(err)
		return
	}
	cResolver = NewConsulResolver(cli2)

	config3 := consulapi.DefaultConfig()
	config3.Address = consulAddr
	cli3, err := consulapi.NewClient(config3)
	if err != nil {
		log.Fatal(err)
		return
	}
	consulClient = cli3
}

// TestNewConsulRegister tests the NewConsulRegister function.
func TestNewConsulRegister(t *testing.T) {
	config := consulapi.DefaultConfig()
	config.Address = consulAddr
	cli, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal(err)
		return
	}
	consulRegister := NewConsulRegister(cli)
	assert.NotNil(t, consulRegister)
}

// TestNewConsulResolver tests the NewConsulResolver function with check option.
func TestNewConsulRegisterWithCheckOption(t *testing.T) {
	config := consulapi.DefaultConfig()
	config.Address = consulAddr
	cli, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal(err)
		return
	}

	check := new(consulapi.AgentServiceCheck)
	check.Timeout = "10s"
	check.Interval = "10s"
	check.DeregisterCriticalServiceAfter = "1m"

	consulResolver := NewConsulResolver(cli, WithCheck(check))
	assert.NotNil(t, consulResolver)
}

// TestNewConsulResolver tests the NewConsulResolver function .
func TestNewConsulResolver(t *testing.T) {
	config := consulapi.DefaultConfig()
	config.Address = consulAddr
	cli, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal(err)
		return
	}

	consulResolver := NewConsulResolver(cli)
	assert.NotNil(t, consulResolver)
}

// TestNewConsulResolver tests unit test preparatory work.
func TestConsulPrepared(t *testing.T) {
	assert.NotNil(t, consulClient)
	assert.NotNil(t, cRegistry)
	assert.NotNil(t, cResolver)
	assert.NotEmpty(t, localIpAddr)
}

// TestRegister tests the Register function.
func TestRegister(t *testing.T) {
	var (
		testSvcName   = "demo.svc.local"
		testSvcPort   = fmt.Sprintf("%d", 8081)
		testSvcWeight = 777
		metaList      = map[string]string{
			"k1": "vv1",
			"k2": "vv2",
			"k3": "vv3",
		}
	)
	// listen on the port, and wait for the health check to connect
	addr := net.JoinHostPort(localIpAddr, testSvcPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		t.Errorf("listen tcp %s failed, err=%v", addr, err)
		t.Fail()
	}
	defer func() {
		if lis != nil {
			lis.Close()
		}
	}()

	testSvcAddr := utils.NewNetAddr("tcp", addr)
	registryInfo := &registry.Info{
		ServiceName: testSvcName,
		Weight:      testSvcWeight,
		Addr:        testSvcAddr,
		Tags:        metaList,
	}
	err = cRegistry.Register(registryInfo)
	assert.Nil(t, err)
	// wait for health check passing
	time.Sleep(time.Second * 6)

	list, _, err := consulClient.Health().Service(testSvcName, "", true, nil)
	assert.Nil(t, err)
	if assert.Equal(t, 1, len(list)) {
		ss := list[0]
		gotSvc := ss.Service
		assert.Equal(t, testSvcName, gotSvc.Service)
		assert.Equal(t, testSvcAddr.String(), net.JoinHostPort(gotSvc.Address, fmt.Sprintf("%d", gotSvc.Port)))
		assert.Equal(t, testSvcWeight, gotSvc.Weights.Passing)
		assert.Equal(t, metaList, gotSvc.Meta)
	}
}

// TestConsulDiscovery tests the ConsulDiscovery function.
func TestConsulDiscovery(t *testing.T) {
	var (
		testSvcName   = strconv.Itoa(int(time.Now().Unix())) + ".svc.local"
		testSvcPort   = fmt.Sprintf("%d", 8082)
		testSvcWeight = 777
		ctx           = context.Background()
	)
	// listen on the port, and wait for the health check to connect
	addr := net.JoinHostPort(localIpAddr, testSvcPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		t.Errorf("listen tcp %s failed, err=%s", addr, err.Error())
		t.Fail()
	}
	defer func() {
		if lis != nil {
			lis.Close()
		}
	}()

	testSvcAddr := utils.NewNetAddr("tcp", addr)
	info := &registry.Info{
		ServiceName: testSvcName,
		Weight:      testSvcWeight,
		Addr:        testSvcAddr,
	}
	err = cRegistry.Register(info)
	assert.Nil(t, err)
	// wait for health check passing
	time.Sleep(time.Second * 6)

	// resolve
	result, err := cResolver.Resolve(ctx, testSvcName)
	assert.Nil(t, err)
	if assert.Equal(t, 1, len(result.Instances)) {
		instance := result.Instances[0]
		assert.Equal(t, testSvcWeight, instance.Weight())
		assert.Equal(t, testSvcAddr.String(), instance.Address().String())
	}
}

// TestDeregister tests the Deregister function.
func TestDeregister(t *testing.T) {
	var (
		testSvcName   = strconv.Itoa(int(time.Now().Unix())) + ".svc.local"
		testSvcPort   = fmt.Sprintf("%d", 8083)
		testSvcWeight = 777
		ctx           = context.Background()
	)

	// listen on the port, and wait for the health check to connect
	addr := net.JoinHostPort(localIpAddr, testSvcPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		t.Errorf("listen tcp %s failed, err=%s", addr, err.Error())
		t.Fail()
	}
	defer func() {
		if lis != nil {
			lis.Close()
		}
	}()

	testSvcAddr := utils.NewNetAddr("tcp", addr)
	info := &registry.Info{
		ServiceName: testSvcName,
		Weight:      testSvcWeight,
		Addr:        testSvcAddr,
	}
	err = cRegistry.Register(info)
	assert.Nil(t, err)
	time.Sleep(time.Second * 6)

	// resolve
	result, err := cResolver.Resolve(ctx, testSvcName)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result.Instances))

	// deregister
	err = cRegistry.Deregister(info)
	assert.Nil(t, err)
	time.Sleep(time.Second)

	// resolve again
	result, err = cResolver.Resolve(ctx, testSvcName)
	assert.NotNil(t, err)
	assert.Equal(t, errors.New("no service found"), err)
}

// TestMultiInstancesRegister tests the Register function, register multiple instances, then deregister one of them.
func TestMultiInstancesRegister(t *testing.T) {
	var (
		testSvcName = "svc.local"

		testSvcPort1 = fmt.Sprintf("%d", 8091)
		testSvcPort2 = fmt.Sprintf("%d", 8092)
		testSvcPort3 = fmt.Sprintf("%d", 8093)
	)

	addr1 := net.JoinHostPort(localIpAddr, testSvcPort1)
	lis1, err := net.Listen("tcp", addr1)
	if err != nil {
		t.Errorf("listen tcp %s failed, err=%s", addr1, err.Error())
		t.Fail()
	}
	defer func() {
		if lis1 != nil {
			lis1.Close()
		}
	}()
	testSvcAddr := utils.NewNetAddr("tcp", addr1)
	err = cRegistry.Register(&registry.Info{
		ServiceName: testSvcName,
		Weight:      11,
		Addr:        testSvcAddr,
	})
	assert.Nil(t, err)

	addr2 := net.JoinHostPort(localIpAddr, testSvcPort2)
	lis2, err := net.Listen("tcp", addr2)
	if err != nil {
		t.Errorf("listen tcp %s failed, err=%s", addr2, err.Error())
		t.Fail()
	}
	defer func() {
		if lis2 != nil {
			lis2.Close()
		}
	}()
	testSvcAddr2 := utils.NewNetAddr("tcp", addr2)
	err = cRegistry.Register(&registry.Info{
		ServiceName: testSvcName,
		Weight:      22,
		Addr:        testSvcAddr2,
	})
	assert.Nil(t, err)

	addr3 := net.JoinHostPort(localIpAddr, testSvcPort3)
	lis3, err := net.Listen("tcp", addr3)
	if err != nil {
		t.Errorf("listen tcp %s failed, err=%s", addr3, err.Error())
		t.Fail()
	}
	defer func() {
		if lis3 != nil {
			lis3.Close()
		}
	}()
	testSvcAddr3 := utils.NewNetAddr("tcp", addr3)
	err = cRegistry.Register(&registry.Info{
		ServiceName: testSvcName,
		Weight:      33,
		Addr:        testSvcAddr3,
	})
	assert.Nil(t, err)

	time.Sleep(time.Second * 6)

	svcList, _, err := consulClient.Health().Service(testSvcName, "", true, nil)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(svcList))

	err = cRegistry.Deregister(&registry.Info{
		ServiceName: testSvcName,
		Weight:      22,
		Addr:        testSvcAddr2,
	})
	assert.Nil(t, err)
	svcList, _, err = consulClient.Health().Service(testSvcName, "", true, nil)
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(svcList)) {
		for _, entry := range svcList {
			gotSvc := entry.Service
			assert.Equal(t, testSvcName, gotSvc.Service)
			assert.Contains(t, []string{testSvcPort1, testSvcPort3}, fmt.Sprintf("%d", gotSvc.Port))
			assert.Equal(t, localIpAddr, gotSvc.Address)
		}
	}
}
