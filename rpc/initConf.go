package rpc

import (
	config2 "github.com/flyerxp/lib/config"
	"github.com/flyerxp/lib/logger"
	ymls "github.com/flyerxp/lib/utils/yaml"
	"go.uber.org/zap"
)

/*
service_name: 'user.info.a'
timeout: 1s #超时设置
max_retry_times: 1 #最大重试次数
eer_threshold: 10%  #重试熔断错误率阈值, 方法级别请求错误率超过阈值则停止重试
#service_find:

	#type: nacos
	#service_conf_name: nacosConf
*/
type ServerStt struct {
	ServiceName string `yaml:"service_name" json:"service_name,omitempty"` //服务名
	Address     string `yaml:"address" json:"address,omitempty"`           //绑定地址
	LogLevel    string `yaml:"log_level" json:"log_level"`
	ServiceFind struct {
		Type string `yaml:"type" json:"type,omitempty"` //服务发现类型
	} `yaml:"service_find" json:"service_find"`
}
type ClientStt struct {
	RpcTimeout     string `yaml:"rpc_timeout" json:"rpc_timeout,omitempty"`         //超时时间
	ConnTimeout    string `yaml:"conn_timeout" json:"conn_timeout"`                 //连接超时
	MaxRetryTimes  int    `yaml:"max_retry_times" json:"max_retry_times,omitempty"` //最大重试次数
	MaxDurationMS  int    `yaml:"max_duration_ms" json:"max_duration_ms,omitempty"` //累计最大耗时，包括首次失败请求和重试请求耗时，如果耗时达到了限制的时间则停止后续的重试。0 表示无限制。注意：如果配置，该配置项必须大于请求超时时间。
	EERThreshold   int    `yaml:"eer_threshold" json:"eer_threshold,omitempty"`     //#重试熔断错误率阈值, 方法级别请求错误率超过阈值则停止重试 10 是指10%
	ConnType       string `yaml:"conn_type" json:"conn_type,omitempty"`             //#使用长连接还是短连接 short|long
	Warmup         bool   `yaml:"warmup" json:"warmup"`
	WarmupConnNums int    `yaml:"warmup_conn_nums" json:"warmup_conn_nums"`
	Pool           struct {
		MaxIdlePerAddress int    `yaml:"max_idle_per_address" json:"max_idle_per_address,omitempty"` //表示每个后端实例可允许的最大闲置连接数
		MaxIdleGlobal     int    `yaml:"max_idle_global" json:"max_idle_global,omitempty"`           //表示全局最大闲置连接数
		MaxIdleTimeout    string `yaml:"max_idle_timeout" json:"max_idle_timeout,omitempty"`         //表示连接的闲置时长，超过这个时长的连接会被关闭（最小值 3s，默认值 30s ）
		MinIdlePerAddress int    `yaml:"min_idle_per_address" json:"min_idle_per_address,omitempty"` //对每个后端实例维护的最小空闲连接数，这部分连接即使空闲时间超过 MaxIdleTimeout 也不会被清理。
	} `yaml:"pool" json:"pool"`
	ServiceFind struct {
		Type string `yaml:"type" json:"type,omitempty"` //服务发现类型
	} `yaml:"service_find" json:"service_find"`
}
type KitexConf struct {
	Kitex struct {
		Server ServerStt        `yaml:"server" json:"server"`
		Client ClientStt        `yaml:"client" json:"client"`
		Nacos  config2.MidNacos `yaml:"nacos" json:"nacos"`
	} `yaml:"kitex" json:"kitex"`
	IsInitEnd bool
}

func GetConf(yaml string) KitexConf {
	KitexConfV := new(KitexConf)
	err := ymls.DecodeByBytes([]byte(yaml), KitexConfV)
	if err != nil {
		logger.ErrWithoutCtx(zap.Error(err), zap.String("rpcConfig", yaml))
	}
	return *KitexConfV
}
