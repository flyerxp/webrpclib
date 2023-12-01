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
	"strings"
)

var serverIp string

func getServerIp() string {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return ""
	}
	addr := conn.LocalAddr().(*net.UDPAddr)
	return strings.Split(addr.String(), ":")[0]
}

// NewServer creates a server.Server with the given handler and options.
func GetServerOptions(sName string, yaml string, opts ...server.Option) []server.Option {
	var options []server.Option
	conf := GetConf(yaml)
	addr, err := net.ResolveTCPAddr("tcp", conf.Kitex.Server.Address)
	if err != nil {
		panic(err)
	}
	serverIp = getServerIp() + ":" + strconv.Itoa(addr.Port)

	if conf.Kitex.Server.Address != "" && conf.Kitex.Server.Address[0:1] != ":" {
		options = append(options, server.WithServiceAddr(addr))
	} else if conf.Kitex.Server.Address[0:1] == ":" {
		addr, err = net.ResolveTCPAddr("tcp", serverIp)
		options = append(options, server.WithServiceAddr(addr))
	}
	if conf.Kitex.Server.ServiceFind.Type != "" {
		options = append(options, getServiceFind(&conf)...)
	}
	options = append(options, server.WithErrorHandler(func(ctx context.Context, err error) error {
		logger.ErrWithoutCtx(zap.String("type", "server"), zap.Error(err))
		return err
	}))
	options = append(options, server.WithMetaHandler(ServerTTHeaderHandler))
	options = append(options, opts...)
	initRpcLog(&conf)
	return options
}

// 获取服务发现配置
func getServiceFind(conf *KitexConf) []server.Option {
	if conf.Kitex.Server.ServiceFind.Type == "nacos" {
		return getServerNacosOption(conf)
	}
	return []server.Option{}
}
func getServerNacosOption(conf *KitexConf) []server.Option {
	option := make([]server.Option, 0, 2)
	cli := getNacosClient(conf)
	option = append(option,
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: conf.Kitex.Server.ServiceName, Tags: map[string]string{}}),
		//server.WithTransHandlerFactory(&remote.ServerTransHandlerFactory{}),
		server.WithRegistry(registry.NewNacosRegistry(cli,
			registry.WithCluster("rpc"),
			registry.WithGroup("rpc"),
		)),
	)
	return option
}
