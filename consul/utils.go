package consul

import (
	"fmt"
	"net"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
)

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

func getServiceId(info *registry.Info) (string, error) {
	host, port, err := parseAddr(info.Addr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s:%d", info.ServiceName, host, port), nil
}
