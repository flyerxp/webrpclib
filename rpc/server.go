package rpc

import (
	"context"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/flyerxp/lib/logger"
	"github.com/kitex-contrib/registry-nacos/registry"
	"go.uber.org/zap"
	"net"
	"strconv"
)

var serverIp string

func getServerIp() string {
	addr, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addr {
		// 检查IP地址，其他类型的地址(如link-local或者loopback)忽略
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return ""
}

// NewServer creates a server.Server with the given handler and options.
func GetServerOptions(sName string, yaml string, opts ...server.Option) []server.Option {
	var options []server.Option
	initConf(yaml)
	conf := GetConf()
	addr, err := net.ResolveTCPAddr("tcp", conf.Kitex.Server.Address)
	if err != nil {
		panic(err)
	}
	serverIp = getServerIp() + ":" + strconv.Itoa(addr.Port)
	if conf.Kitex.Server.Address != "" {
		options = append(options, server.WithServiceAddr(addr))
	}
	if conf.Kitex.Server.ServiceFind.Type != "" {
		options = append(options, getServiceFind(sName)...)
	}
	options = append(options, server.WithErrorHandler(func(ctx context.Context, err error) error {
		logger.ErrWithoutCtx(zap.String("type", "server"), zap.Error(err))
		return err
	}))

	options = append(options, server.WithMetaHandler(ServerTTHeaderHandler))
	options = append(options, opts...)
	initRpcLog()
	return options
}

// 获取服务发现配置
func getServiceFind(sName string) []server.Option {
	conf := GetConf()
	if conf.Kitex.Server.ServiceFind.Type == "nacos" {
		return getServerNacosOption(sName)
	}
	return []server.Option{}
}
func getServerNacosOption(sName string) []server.Option {
	option := make([]server.Option, 0, 2)
	cli := getNacosClient()
	option = append(option,
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: sName, Tags: map[string]string{}}),
		server.WithRegistry(registry.NewNacosRegistry(cli,
			registry.WithCluster("rpc"),
			registry.WithGroup("rpc"),
		)),
	)
	return option
}
