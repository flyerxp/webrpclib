package rpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"github.com/cloudwego/kitex/client"
	cpp "github.com/cloudwego/kitex/pkg/connpool"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/remote/codec/protobuf"
	cp "github.com/cloudwego/kitex/pkg/remote/connpool"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/grpc"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/status"
	"github.com/cloudwego/kitex/pkg/retry"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/warmup"
	config2 "github.com/flyerxp/lib/v2/config"
	"github.com/flyerxp/lib/v2/logger"
	"github.com/kitex-contrib/registry-nacos/resolver"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"go.uber.org/zap"
	"net"
	"net/url"
	"strconv"
	"time"
)

type nacosClientStt struct {
	RpcClient naming_client.INamingClient
	IsInitEnd bool
}

var nacosClient *nacosClientStt

func init() {
	nacosClient = new(nacosClientStt)
}

type ConnReporter struct {
}

func (c *ConnReporter) ConnSucceed(poolType cp.ConnectionPoolType, serviceName string, addr net.Addr) {
	L.WithField("service", serviceName).
		WithField("addr", addr.String()).
		WithField("poolType", poolType).
		Info("Succeed")
}
func (c *ConnReporter) ConnFailed(poolType cp.ConnectionPoolType, serviceName string, addr net.Addr) {
	L.WithField("service", serviceName).
		WithField("addr", addr.String()).
		WithField("poolType", poolType).
		Warn("Failed")
}
func (c *ConnReporter) ReuseSucceed(poolType cp.ConnectionPoolType, serviceName string, addr net.Addr) {

}

// NewClientOption
func GetClientOptions(yaml string, opts ...client.Option) []client.Option {
	conf := GetConf(yaml)
	ObjConnPool := getPool(&conf)
	re := new(ConnReporter)
	cp.SetReporter(re)
	options := []client.Option{
		client.WithConnReporterEnabled(),
		client.WithCloseCallbacks(func() error {
			logger.WarnWithoutCtx(zap.String("client", "close"))
			return nil
		}),
		client.WithErrorHandler(func(ctx context.Context, err error) error {
			switch err.(type) {
			case *remote.TransError, thrift.TApplicationException, protobuf.PBError:
				logger.ErrWithoutCtx(zap.String("RpicClientIo", err.Error()), zap.Error(err))
				return kerrors.ErrRemoteOrNetwork.WithCauseAndExtraMsg(err, "remote")
			}
			logger.ErrWithoutCtx(zap.String("RpcRemoteOrNetwork", err.Error()), zap.Error(err))
			return kerrors.ErrRemoteOrNetwork.WithCause(err)
		}),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: config2.GetConf().App.Name,
		}),
		client.WithMetaHandler(ClientTTHeaderHandler),
	}
	if conf.Kitex.Client.ConnType == "short" {
		options = append(options, client.WithShortConnection())
	} else {
		options = append(options, client.WithLongConnection(ObjConnPool))
	}
	if conf.Kitex.Client.ServiceFind.Type != "" {
		options = append(options, getClientNacosOption(&conf))
	}
	fp := retry.NewFailurePolicy()

	if conf.Kitex.Client.MaxRetryTimes > 0 {
		fp.WithMaxRetryTimes(conf.Kitex.Client.MaxRetryTimes) // 配置最多重试3次
	}
	if conf.Kitex.Client.EERThreshold > 0 {
		fp.WithRetryBreaker(float64(conf.Kitex.Client.EERThreshold) / 100)
	}
	if conf.Kitex.Client.MaxDurationMS > 0 {
		fp.WithMaxDurationMS(uint32(conf.Kitex.Client.MaxDurationMS))
	}

	if conf.Kitex.Client.KeepLive.PermitWithoutStream && conf.Kitex.Client.KeepLive.Time != "" {
		t, errT := time.ParseDuration(conf.Kitex.Client.KeepLive.Time)
		if errT != nil {
			panic(errors.New(fmt.Sprintf("conf.Kitex.Client.KeepLive.Time 解析错误:%s", conf.Kitex.Client.KeepLive.Time)))
		}
		to, errTo := time.ParseDuration(conf.Kitex.Client.KeepLive.TimeOut)
		if errTo != nil {
			panic(errors.New(fmt.Sprintf("conf.Kitex.Client.KeepLive.TimeOut 解析错误:%s", conf.Kitex.Client.KeepLive.TimeOut)))
		}
		options = append(options, client.WithGRPCKeepaliveParams(grpc.ClientKeepalive{
			Time:                t, // less than  grpc.KeepaliveMinPingTime(5s)
			Timeout:             to,
			PermitWithoutStream: true,
		}))
	}
	options = append(options, client.WithFailureRetry(fp))
	if conf.Kitex.Client.RpcTimeout != "" {
		t, errT := time.ParseDuration(conf.Kitex.Client.RpcTimeout)
		if errT != nil {
			panic(errT)
		}
		options = append(options, client.WithRPCTimeout(t))
	}
	if conf.Kitex.Client.ConnTimeout != "" {
		t, errT := time.ParseDuration(conf.Kitex.Client.ConnTimeout)
		if errT != nil {
			panic(errT)
		}
		options = append(options, client.WithRPCTimeout(t))
	}

	if conf.Kitex.Client.Warmup {
		options = append(options, client.WithWarmingUp(&warmup.ClientOption{
			PoolOption: &warmup.PoolOption{
				ConnNum: conf.Kitex.Client.WarmupConnNums,
			},
		}))
	}

	options = append(options, opts...)
	return options
}

func getClientNacosOption(conf *KitexConf) client.Option {
	cli := getNacosClient(conf)
	return client.WithResolver(resolver.NewNacosResolver(cli,
		resolver.WithCluster("rpc"),
		resolver.WithGroup("rpc"),
	))
}
func getNacosClient(conf *KitexConf) naming_client.INamingClient {
	if nacosClient.IsInitEnd == false {
		c := conf.Kitex.Nacos
		oUrl, _ := url.Parse(c.Url)
		iPort, _ := strconv.Atoi(oUrl.Port())
		sc := []constant.ServerConfig{
			*constant.NewServerConfig(oUrl.Hostname(), uint64(iPort)),
		}
		cc := constant.ClientConfig{
			NamespaceId:         c.Ns,
			TimeoutMs:           5000,
			NotLoadCacheAtStart: true,
			LogDir:              "logs/nacos/logs",
			CacheDir:            "logs/nacos/cache",
			LogLevel:            "warn",
			Username:            c.User,
			Password:            c.Pwd,
		}
		cli, err := clients.NewNamingClient(
			vo.NacosClientParam{
				ClientConfig:  &cc,
				ServerConfigs: sc,
			},
		)
		if err != nil {
			panic(err)
		}
		nacosClient.RpcClient = cli
		nacosClient.IsInitEnd = true
	}
	return nacosClient.RpcClient
}

// connpool init
func getPool(conf *KitexConf) cpp.IdleConfig {
	ObjConnPool := cpp.IdleConfig{MaxIdlePerAddress: 10, MaxIdleGlobal: 1000, MaxIdleTimeout: time.Minute, MinIdlePerAddress: 2}

	pollConf := conf.Kitex.Client.Pool
	if pollConf.MaxIdlePerAddress > 0 {
		ObjConnPool.MaxIdlePerAddress = pollConf.MaxIdlePerAddress
	}
	if pollConf.MaxIdleGlobal > 0 {
		ObjConnPool.MaxIdleGlobal = pollConf.MaxIdleGlobal
	}
	if pollConf.MaxIdleTimeout != "" {
		t, e := time.ParseDuration(pollConf.MaxIdleTimeout)
		if e != nil {
			logger.ErrWithoutCtx(zap.Error(e))
		} else {
			ObjConnPool.MaxIdleTimeout = t
		}
	}
	if pollConf.MinIdlePerAddress > 0 {
		ObjConnPool.MinIdlePerAddress = pollConf.MinIdlePerAddress
	}

	return ObjConnPool
}
func RecordError(ctx context.Context, err error, field ...zap.Field) {
	field = append(field, zap.Error(err))
	logger.AddError(ctx, field...)
	if e, ok := err.(*remote.TransError); ok {
		logger.AddError(ctx, zap.String("trans", e.Error()))
		klog.CtxErrorf(ctx, "NewsView trans call failed,err =%+v", err)
		logger.WriteErr(ctx)
	}
	if s, ok := status.FromError(err); ok {
		logger.AddError(ctx, zap.String("rpcMessage", s.Message()))
		klog.CtxErrorf(ctx, "NewsView rpc call failed,err =%+v", err)
		logger.WriteErr(ctx)
	}
}
