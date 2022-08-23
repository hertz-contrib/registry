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

// NewEurekaResolver creates a eureka resolver.
func NewEurekaResolver(servers []string) discovery.Resolver {
	conn := fargo.NewConn(servers...)

	return &eurekaResolver{eurekaConn: &conn}
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
