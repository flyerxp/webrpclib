package web

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// nacosRegistry 基于 nacos-sdk-go/v2 naming 客户端实现 hertz registry.Registry。
type nacosRegistry struct {
	cli     naming_client.INamingClient
	cluster string
	group   string
}

var _ registry.Registry = (*nacosRegistry)(nil)

func newNacosRegistry(cli naming_client.INamingClient, cluster, group string) registry.Registry {
	return &nacosRegistry{cli: cli, cluster: cluster, group: group}
}

// Register service info to nacos.
func (n *nacosRegistry) Register(info *registry.Info) error {
	if err := validateRegistryInfo(info); err != nil {
		return fmt.Errorf("valid parse registry info error: %w", err)
	}
	host, port, err := net.SplitHostPort(info.Addr.String())
	if err != nil {
		return fmt.Errorf("parse registry info addr error: %w", err)
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("parse registry info port error: %w", err)
	}
	if host == "" || host == "::" {
		host = utils.LocalIP()
	}
	if _, err = n.cli.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          host,
		Port:        uint64(p),
		ServiceName: info.ServiceName,
		GroupName:   n.group,
		ClusterName: n.cluster,
		Weight:      float64(info.Weight),
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    info.Tags,
	}); err != nil {
		return fmt.Errorf("register instance error: %w", err)
	}
	return nil
}

// Deregister a service info from nacos.
func (n *nacosRegistry) Deregister(info *registry.Info) error {
	if err := validateRegistryInfo(info); err != nil {
		return fmt.Errorf("valid parse registry info error: %w", err)
	}
	host, port, err := net.SplitHostPort(info.Addr.String())
	if err != nil {
		return err
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("parse registry info port error: %w", err)
	}
	if host == "" || host == "::" {
		host = utils.LocalIP()
	}
	if _, err = n.cli.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          host,
		Port:        uint64(p),
		ServiceName: info.ServiceName,
		GroupName:   n.group,
		Cluster:     n.cluster,
		Ephemeral:   true,
	}); err != nil {
		return err
	}
	return nil
}

func validateRegistryInfo(info *registry.Info) error {
	if info == nil {
		return errors.New("registry.Info can not be empty")
	}
	if info.ServiceName == "" {
		return errors.New("registry.Info ServiceName can not be empty")
	}
	if info.Addr == nil {
		return errors.New("registry.Info Addr can not be empty")
	}
	return nil
}
