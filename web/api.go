package web

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/app/server/render"
	"github.com/cloudwego/hertz/pkg/common/config"
	config2 "github.com/flyerxp/lib/config"
	"github.com/flyerxp/lib/logger"
	"github.com/flyerxp/lib/utils/json"
	htt2conf "github.com/hertz-contrib/http2/config"
	"github.com/hertz-contrib/http2/factory"
	"github.com/hertz-contrib/logger/accesslog"
	"github.com/hertz-contrib/pprof"
	"github.com/hertz-contrib/registry/nacos"
	"github.com/hertz-contrib/requestid"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"go.uber.org/zap"
	"net/url"
	"strconv"
	"time"
)

func GetHttpServer(ctx context.Context, rInfo *registry.Info) *server.Hertz {
	heConf := GetConf()
	idle, err1 := time.ParseDuration(heConf.Hertz.IdleTimeout)
	read, err2 := time.ParseDuration(heConf.Hertz.ReadTimeout)
	write, err3 := time.ParseDuration(heConf.Hertz.WriteTimeout)
	if err1 != nil {
		logger.AddError(ctx, zap.Error(err1))
		panic(err1)
	}
	if err2 != nil {
		logger.AddError(ctx, zap.Error(err2))
		panic(err2)
	}
	if err3 != nil {
		logger.AddError(ctx, zap.Error(err3))
		panic(err3)
	}
	confOption := []config.Option{
		server.WithIdleTimeout(idle),
		server.WithReadTimeout(read),
		server.WithMaxRequestBodySize(heConf.Hertz.MaxRequestBodySize),
		server.WithRemoveExtraSlash(true),
		server.WithWriteTimeout(write),
		server.WithH2C(heConf.Hertz.H2c),
	}
	strAddress := heConf.Hertz.Address
	strAddress += ":" + string(heConf.Hertz.Port)
	confOption = append(confOption, server.WithHostPorts(strAddress))
	//服务发现
	if heConf.Hertz.ServiceFind.ServiceConfName != "" && heConf.Hertz.ServiceFind.Type != "" {
		//目前只支持nacos
		if heConf.Hertz.ServiceFind.Type == "nacos" {
			confOption = append(confOption, getNacosOption(heConf.Hertz.ServiceFind.ServiceConfName, rInfo))
		}
	}
	h := server.Default(confOption...)
	h.Name = config2.GetConf().App.Name
	registerMiddleware(heConf, h, rInfo)
	render.ResetJSONMarshal(json.Encode)
	return h
}

func registerMiddleware(heConf *HertzConf, h *server.Hertz, rInfo *registry.Info) {
	h.Use(requestid.New())
	h.Use(func(ctx context.Context, c *app.RequestContext) {
		ctx = logger.GetContext(ctx, fmt.Sprintf("web_%s", requestid.Get(c)))
		c.Next(ctx)
	})
	if heConf.Hertz.EnableAccessLog {
		h.Use(
			accesslog.New(accesslog.WithTimeInterval(time.Second),
				accesslog.WithAccessLogFunc(logger.WriteAccess),
				accesslog.WithFormat("{\"time\":\"${time}\",\"status\":\"${status}\",\"latency\":\"${latency}\",\"method\":\"${method}\",\"path:\"\"${path}\",\"query\":\"${queryParams}\"}"),
			))
	}
	if heConf.Hertz.EnablePprof {
		pprof.Register(h)
	}
	if heConf.Hertz.H2c {
		h.AddProtocol("h2", factory.NewServerFactory(
			htt2conf.WithReadTimeout(time.Minute),
			htt2conf.WithDisableKeepAlive(false)))
	}
	ip := rInfo.Addr.String()
	h.Use(func(ctx context.Context, c *app.RequestContext) {
		logger.SetUrl(ctx, fmt.Sprintf("%s?%s", string(c.Path()), c.QueryArgs().String()))
		if string(c.ContentType()) == "application/json" {
			logger.SetArgs(ctx, string(c.GetRequest().Body()))
		} else {
			logger.SetArgs(ctx, c.PostArgs().String())
		}
		logger.SetRefer(ctx, string(c.GetHeader("Referer")))
		c.Set("loggerStart", time.Now())
		c.Next(ctx)
	}, func(ctx context.Context, c *app.RequestContext) {
		c.Next(ctx)
		logger.SetAddr(ctx, c.ClientIP(), ip)
		logger.SetUserAgent(ctx, string(c.UserAgent()))
		if v, ok := c.Get("loggerStart"); ok {
			logger.SetExecTime(ctx, int(time.Since(v.(time.Time)).Microseconds()))
		}
		logger.WriteLine(ctx)
	})

}
func getNacosOption(sName string, rInfo *registry.Info) config.Option {
	nConf := config2.MidNacos{}
	for _, v := range config2.GetConf().Nacos {
		if v.Name == sName {
			nConf = v
			break
		}
	}
	if nConf.Url == "" {
		panic(errors.New("no find nocas config Name:" + sName))
	}
	sf := getNacosFind(nConf)
	return server.WithRegistry(sf, rInfo)
}

// 获取服务发现配置
func getNacosFind(c config2.MidNacos) registry.Registry {
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
		AppName:             config2.GetConf().App.Name,
	}

	cli, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	r := nacos.NewNacosRegistry(cli, nacos.WithRegistryCluster("web"), nacos.WithRegistryGroup("web"))
	if err != nil {
		panic(err)
		return nil
	}
	return r
}
