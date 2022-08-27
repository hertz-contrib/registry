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

package main

import (
	"context"
	"log"

	hclient "github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/hertz-contrib/registry/polaris"
)

const (
	Namespace = "default"
	// At present,polaris server tag is v1.4.0，can't support auto create namespace,
	// if you want to use a namespace other than default,Polaris ,before you register an instance,
	// you should create the namespace at polaris console first.
)

func main() {
	r, err := polaris.NewPolarisResolver()
	if err != nil {
		log.Fatal(err)
	}

	client, err := hclient.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	client.Use(sd.Discovery(r))

	for i := 0; i < 10; i++ {
		// config.WithTag sets the namespace tag for service discovery
		status, body, err := client.Get(context.TODO(), nil, "http://hertz.test.demo/hello", config.WithSD(true), config.WithTag("namespace", Namespace))
		if err != nil {
			hlog.Fatal(err)
		}
		hlog.Infof("code=%d,body=%s\n", status, body)
	}
}
