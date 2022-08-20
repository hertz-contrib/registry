package eureka

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/hudl/fargo"
	"net"
	"strconv"
	"time"
)

type eurekaRegistry struct {
	HeatBeatInterval   time.Duration
	convInfoToInstance func(info *registry.Info) (*fargo.Instance, error)
	conn               fargo.EurekaConnection
	instance           *fargo.Instance
	heartBeatCancel    context.CancelFunc
}

func NewEurekaRegister(cli fargo.EurekaConnection) registry.Registry {

	return &eurekaRegistry{
		conn:     cli,
		instance: nil,
	}
}

func (c eurekaRegistry) CheckIsInstanceNil() error {
	if c.instance == nil {
		return fmt.Errorf("eureka instance is nil. is this instance registered")
	}

	return nil
}

func (c eurekaRegistry) setInstance(ins *fargo.Instance) {
	c.instance = ins
}

func (c eurekaRegistry) resetInstance() {
	c.instance = nil
}

func (c eurekaRegistry) sendHeartBeat(ctx context.Context) error {
	if err := c.CheckIsInstanceNil(); err != nil {
		return err
	}

	heartBeatTicker := time.Tick(c.HeatBeatInterval)

	// what to do if heartbeat failed multiple times in a row
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-heartBeatTicker:
			return c.conn.HeartBeatInstance(c.instance)
		}
	}
}

func (c eurekaRegistry) startHeartbeat() error {

	ctx, cancel := context.WithCancel(context.Background())
	c.heartBeatCancel = cancel

	go func() {
		defer func() {
			if err := recover(); err != nil {

			}
		}()
		err := c.sendHeartBeat(ctx)
		if err != nil {
			return
		}
	}()

	return nil
}

func (c eurekaRegistry) Register(info *registry.Info) error {

	instance, err := c.convInfoToInstance(info)
	if err != nil {
		return err
	}

	err = c.conn.RegisterInstance(c.instance)
	if err != nil {
		return err
	}
	c.setInstance(instance)

	err = c.startHeartbeat()
	if err != nil {
		return err
	}

	return nil

}

func (c eurekaRegistry) Deregister(info *registry.Info) error {
	if err := c.CheckIsInstanceNil(); err != nil {
		return err
	}
	err := c.conn.DeregisterInstance(c.instance)
	if err != nil {
		return fmt.Errorf("failed to deregister instance: %v, error: %w", c.instance, err)
	}

	// stop sending heart beat to eureka server
	c.heartBeatCancel()

	// clear registered instance
	c.resetInstance()

	return nil

}

func parseAddr(addr net.Addr) (host string, port int, err error) {
	host, portStr, err := net.SplitHostPort(addr.String())
	if err != nil {
		return "", 0, fmt.Errorf("calling net.SplitHostPort failed, addr: %s, err: %w", addr.String(), err)
	}

	if host == "" || host == "::" {
		host = utils.LocalIP()
	}

	port, err = strconv.Atoi(portStr)
	if err != nil {
		return "", 0, fmt.Errorf("parsing registry info port failed, portStr:%s, err: %w", portStr, err)
	}
	if port == 0 {
		return "", 0, fmt.Errorf("invalid port %d", port)
	}

	return host, port, nil
}

func getInstanceID(info *registry.Info) (string, error) {
	host, port, err := parseAddr(info.Addr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s:%d", info.ServiceName, host, port), nil
}

func DefaultConvInfoToInstance(info *registry.Info) (*fargo.Instance, error) {

	addr := info.Addr

	instanceID, err := getInstanceID(info)
	if err != nil {
		return nil, fmt.Errorf("getting instance id failed, err: %w", err)
	}

	host, port, err := parseAddr(addr)
	if err != nil {
		return nil, fmt.Errorf("calling net.SplitHostPort failed, addr: %s, err: %w", addr.String(), err)
	}

	metadata, err := xml.Marshal(addr)
	if err != nil {
		return nil, fmt.Errorf("marshall tags failed, tags: %v,  err: %w", info.Tags, err)
	}

	return &fargo.Instance{
		InstanceId:       instanceID,
		HostName:         host,
		Port:             port,
		App:              info.ServiceName,
		IPAddr:           utils.LocalIP(),
		VipAddress:       host,
		SecureVipAddress: host,
		Status:           fargo.UP,
		Metadata:         fargo.InstanceMetadata{Raw: metadata},
		PortEnabled:      true,
	}, nil
}
