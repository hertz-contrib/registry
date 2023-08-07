/*
 * Copyright 2021 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package polaris

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/stretchr/testify/assert"
)

const (
	serviceName = "registry-test"
	namespace   = "basic"
	address     = "127.0.0.1:8888"
)

// TestPolarisRegistryWithHertz test polaris registry complete workflow(service registry|service de-registry|service resolver)with hertz.
func TestPolarisRegistryWithHertz(t *testing.T) {
	r, err := NewPolarisRegistry()
	if err != nil {
		t.Fatal(err)
	}

	Info := &registry.Info{
		ServiceName: serviceName,
		Addr:        utils.NewNetAddr("tcp", address),
		Tags: map[string]string{
			"namespace": namespace,
		},
	}
	h := server.Default(server.WithHostPorts(address), server.WithRegistry(r, Info), server.WithExitWaitTime(10*time.Second))

	h.GET("/hello", func(ctx context.Context, c *app.RequestContext) {
		c.String(consts.StatusOK, "Hello,Hertz!")
	})

	go h.Spin()

	time.Sleep(5 * time.Second) // wait server start

	hclient, err := client.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := NewPolarisResolver()
	if err != nil {
		t.Fatalf("HERTZ: Init polaris resolver error=%v", err)
	}
	hclient.Use(sd.Discovery(resolver))

	uri := fmt.Sprintf("http://%s/hello", serviceName)

	status, body, err := hclient.Get(context.TODO(), nil, uri, config.WithSD(true), config.WithTag("namespace", namespace))

	assert.Nil(t, err)
	assert.Equal(t, 200, status)
	assert.Equal(t, "Hello,Hertz!", string(body))

	if err := h.Shutdown(context.Background()); err != nil {
		t.Errorf("HERTZ: Shutdown error=%v", err)
	}

	time.Sleep(5 * time.Second) // wait server shutdown

	status, body, err = hclient.Get(context.TODO(), nil, uri, config.WithSD(true), config.WithTag("namespace", namespace))

	assert.NotNil(t, err)
	assert.Equal(t, 0, status)
	assert.Equal(t, "", string(body))
}

func TestEmptyEndpoints(t *testing.T) {
	_, err := NewPolarisResolver()
	assert.Nil(t, err)
}
