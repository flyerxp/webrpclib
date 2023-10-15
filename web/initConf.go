package web

import (
	config2 "github.com/flyerxp/lib/config"
	"github.com/flyerxp/lib/logger"
	ymls "github.com/flyerxp/lib/utils/yaml"
	"go.uber.org/zap"
)

type HertzConf struct {
	Hertz struct {
		Address            string `yaml:"address"`
		Port               string `yaml:"port"`
		IdleTimeout        string `yaml:"idle_timeout"`
		ReadTimeout        string `yaml:"read_timeout"`
		MaxRequestBodySize int    `yaml:"max_request_body_size"`
		WriteTimeout       string `yaml:"write_timeout"`
		H2c                bool   `yaml:"h2C"`
		EnablePprof        bool   `yaml:"enable_pprof"`
		EnableAccessLog    bool   `yaml:"enable_access_log"`
		ServiceFind        struct {
			Type            string `yaml:"type"`
			ServiceConfName string `yaml:"service_conf_name"`
		} `yaml:"service_find"`
	} `json:"hertz" yaml:"hertz"`
	Resource struct {
		RootPathImg    string            `yaml:"root_path_img"`
		RootPathJs     string            `yaml:"root_path_js"`
		RootPathCss    string            `yaml:"root_path_css"`
		RootPathNodeJs string            `yaml:"root_path_nodejs"`
		RootHttpUpload string            `yaml:"root_http_upload"`
		RootUpload     string            `yaml:"root_upload"`
		Other          map[string]string `yaml:"other"`
	} `yaml:"resource"`
	IsInitEnd bool
}

var hertzConfV = new(HertzConf)

// 获取配置
func GetConf() *HertzConf {
	if hertzConfV.IsInitEnd == false {
		err := ymls.DecodeByFile(config2.GetConfFile("hertz.yml"), hertzConfV)
		if err != nil {
			logger.ErrWithoutCtx(zap.Error(err), zap.String("file", config2.GetConfFile("hertz.yml")))
		}
		if hertzConfV.Hertz.IdleTimeout == "" {
			hertzConfV.Hertz.IdleTimeout = "60s"
		}
		if hertzConfV.Hertz.ReadTimeout == "" {
			hertzConfV.Hertz.ReadTimeout = "10s"
		}
		if hertzConfV.Hertz.WriteTimeout == "" {
			hertzConfV.Hertz.WriteTimeout = "30s"
		}
		hertzConfV.IsInitEnd = true
	}
	return hertzConfV
}
