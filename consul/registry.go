package consul

import (
	"errors"
	"fmt"
	"net"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/hashicorp/consul/api"
)

type consulRegistry struct {
	consulClient *api.Client
	opts         options
}

var _ registry.Registry = (*consulRegistry)(nil)

type options struct {
	check *api.AgentServiceCheck
}

// Option is the option of Consul.
type Option func(o *options)

// WithCheck is consul registry option to set AgentServiceCheck.
func WithCheck(check *api.AgentServiceCheck) Option {
	return func(o *options) { o.check = check }
}

// NewConsulRegister create a new registry using consul.
func NewConsulRegister(consulClient *api.Client, opts ...Option) registry.Registry {
	op := options{
		check: defaultCheck(),
	}

	for _, opt := range opts {
		opt(&op)
	}

	return &consulRegistry{consulClient: consulClient, opts: op}
}

// Register register a service to consul.
func (c consulRegistry) Register(info *registry.Info) error {
	err := validateRegistryInfo(info)
	if err != nil {
		return fmt.Errorf("validating registry info failed, err: %w", err)
	}

	host, port, err := parseAddr(info.Addr)
	if err != nil {
		return fmt.Errorf("parsing addr failed, err: %w", err)
	}

	svcID, err := getServiceId(info)
	if err != nil {
		return fmt.Errorf("getting service id failed, err: %w", err)
	}

	svcInfo := &api.AgentServiceRegistration{
		ID:      svcID,
		Name:    info.ServiceName,
		Address: host,
		Port:    port,
		Meta:    info.Tags,
		Weights: &api.AgentWeights{
			Passing: info.Weight,
			Warning: info.Weight,
		},
		Check: c.opts.check,
	}
	if c.opts.check != nil {
		c.opts.check.TCP = net.JoinHostPort(host, fmt.Sprintf("%d", port))
		svcInfo.Check = c.opts.check
	}

	return c.consulClient.Agent().ServiceRegister(svcInfo)
}

// Deregister deregister a service from consul.
func (c consulRegistry) Deregister(info *registry.Info) error {
	svcID, err := getServiceId(info)
	if err != nil {
		return err
	}

	return c.consulClient.Agent().ServiceDeregister(svcID)
}

func defaultCheck() *api.AgentServiceCheck {
	check := new(api.AgentServiceCheck)
	check.Timeout = "5s"
	check.Interval = "5s"
	check.DeregisterCriticalServiceAfter = "1m"

	return check
}

func validateRegistryInfo(info *registry.Info) error {
	if info == nil {
		return errors.New("info is nil")
	}
	if info.ServiceName == "" {
		return errors.New("missing service name in consul register")
	}
	if info.Addr == nil {
		return errors.New("missing addr in consul register")
	}

	return nil
}
