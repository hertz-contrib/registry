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

package nacos

import (
	"context"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

var _ discovery.Resolver

type nacosResolver struct {
	client naming_client.INamingClient
	opts   resolverOptions
}

func (n *nacosResolver) Target(_ context.Context, target *discovery.TargetInfo) string {
	var metadata strings.Builder

	// Set serviceName and metadata to desc
	tags := target.Tags
	if len(tags) == 0 {
		return target.Host
	}

	metadata.WriteString(target.Host)
	metadata.WriteString("?")
	values := url.Values{}
	for k, v := range tags {
		values.Add(k, v)
	}
	metadata.WriteString(values.Encode())
	return metadata.String()
}

func (n *nacosResolver) Resolve(_ context.Context, desc string) (discovery.Result, error) {
	var metadata map[string]string
	serviceName := desc

	// Get serviceName and metadata from desc
	if strings.Contains(desc, "?") {
		queries, _ := url.Parse(desc)
		tags, _ := url.ParseQuery(queries.Query().Encode())

		result := make(map[string]string)
		for key, value := range tags {
			result[key] = value[0]
		}
		metadata = result
		serviceName = strings.Split(desc, "?")[0]
	}

	res, err := n.client.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		HealthyOnly: true,
		GroupName:   n.opts.group,
		Clusters:    []string{n.opts.cluster},
	})
	if err != nil {
		return discovery.Result{}, err
	}
	instances := make([]discovery.Instance, 0, len(res))
	for _, ins := range res {
		if !ins.Enable || (len(metadata) > 0 && !compareMaps(ins.Metadata, metadata)) {
			continue
		}

		formatPort := strconv.FormatUint(ins.Port, 10)
		instances = append(instances,
			discovery.NewInstance(
				"tcp",
				net.JoinHostPort(ins.Ip, formatPort),
				int(ins.Weight), ins.Metadata,
			),
		)
	}

	return discovery.Result{
		CacheKey:  desc,
		Instances: instances,
	}, nil
}

func (n *nacosResolver) Name() string {
	return "nacos" + ":" + n.opts.cluster + ":" + n.opts.group
}

// NewDefaultNacosResolver create a default service resolver using nacos.
func NewDefaultNacosResolver(opts ...ResolverOption) (discovery.Resolver, error) {
	client, err := newDefaultNacosConfig()
	if err != nil {
		return nil, err
	}
	return NewNacosResolver(client, opts...), nil
}

// NewNacosResolver create a service resolver using nacos.
func NewNacosResolver(cli naming_client.INamingClient, opts ...ResolverOption) discovery.Resolver {
	opt := resolverOptions{
		cluster: "DEFAULT",
		group:   "DEFAULT_GROUP",
	}
	for _, option := range opts {
		option(&opt)
	}
	return &nacosResolver{client: cli, opts: opt}
}
