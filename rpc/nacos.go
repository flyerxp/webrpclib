package rpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// nacosResolver 基于 nacos-sdk-go/v2 naming 客户端实现 kitex discovery.Resolver。
type nacosResolver struct {
	cli     naming_client.INamingClient
	cluster string
	group   string
}

var _ discovery.Resolver = (*nacosResolver)(nil)

func newNacosResolver(cli naming_client.INamingClient, cluster, group string) discovery.Resolver {
	return &nacosResolver{cli: cli, cluster: cluster, group: group}
}

// Target return a description for the given target that is suitable for being a key for cache.
func (n *nacosResolver) Target(_ context.Context, target rpcinfo.EndpointInfo) (description string) {
	return target.ServiceName()
}

// Resolve a service info by desc.
func (n *nacosResolver) Resolve(_ context.Context, desc string) (discovery.Result, error) {
	res, err := n.cli.SelectInstances(vo.SelectInstancesParam{
		ServiceName: desc,
		HealthyOnly: true,
		GroupName:   n.group,
		Clusters:    []string{n.cluster},
	})
	if err != nil {
		return discovery.Result{}, err
	}
	instances := make([]discovery.Instance, 0, len(res))
	for _, in := range res {
		if !in.Enable {
			continue
		}
		instances = append(instances, discovery.NewInstance(
			"tcp",
			fmt.Sprintf("%s:%d", in.Ip, in.Port),
			int(in.Weight),
			in.Metadata),
		)
	}
	if len(instances) == 0 {
		return discovery.Result{}, fmt.Errorf("no instance remains for %v", desc)
	}
	return discovery.Result{
		Cacheable: true,
		CacheKey:  desc,
		Instances: instances,
	}, nil
}

// Diff computes the difference between two results.
func (n *nacosResolver) Diff(cacheKey string, prev, next discovery.Result) (discovery.Change, bool) {
	return discovery.DefaultDiff(cacheKey, prev, next)
}

// Name returns the name of the resolver.
func (n *nacosResolver) Name() string {
	return "nacos" + ":" + n.cluster + ":" + n.group
}

// nacosRegistry 基于 nacos-sdk-go/v2 naming 客户端实现 kitex registry.Registry。
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
		return err
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
		host, err = getLocalIpv4Host()
		if err != nil {
			return fmt.Errorf("parse registry info addr error: %w", err)
		}
	}
	if _, err = n.cli.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          host,
		Port:        uint64(p),
		ServiceName: info.ServiceName,
		Weight:      float64(info.Weight),
		Enable:      true,
		Healthy:     true,
		Metadata:    info.Tags,
		GroupName:   n.group,
		ClusterName: n.cluster,
		Ephemeral:   true,
	}); err != nil {
		return fmt.Errorf("register instance error: %w", err)
	}
	return nil
}

// Deregister a service info from nacos.
func (n *nacosRegistry) Deregister(info *registry.Info) error {
	if err := validateRegistryInfo(info); err != nil {
		return err
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
		host, err = getLocalIpv4Host()
		if err != nil {
			return fmt.Errorf("parse registry info addr error: %w", err)
		}
	}
	if _, err = n.cli.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          host,
		Port:        uint64(p),
		ServiceName: info.ServiceName,
		Ephemeral:   true,
		GroupName:   n.group,
		Cluster:     n.cluster,
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

func getLocalIpv4Host() (string, error) {
	addr, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addr {
		ipNet, isIpNet := addr.(*net.IPNet)
		if isIpNet && !ipNet.IP.IsLoopback() {
			ipv4 := ipNet.IP.To4()
			if ipv4 != nil {
				return ipv4.String(), nil
			}
		}
	}
	return "", errors.New("not found ipv4 address")
}
