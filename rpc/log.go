package rpc

import (
	"github.com/cloudwego/kitex/pkg/klog"
	config2 "github.com/flyerxp/lib/v2/config"
	"github.com/flyerxp/lib/v2/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	L       *zap.SugaredLogger
	logOnce sync.Once
)

func getLog() *zap.SugaredLogger {
	logOnce.Do(func() {
		L = newZapLogger()
	})
	return L
}

func newZapLogger() *zap.SugaredLogger {
	paths := logger.GetPath(config2.GetConf().App.Logger.OutputPaths, "rpc")
	logPath := ""
	for _, v := range paths {
		if strings.Contains(v, "/") {
			logPath = filepath.Dir(v) + "/rpc_conn.log"
		}
	}

	// 输出目标：控制台 / 文件滚动，参数与原配置完全一致
	var writeSyncer zapcore.WriteSyncer
	if logPath == "" {
		writeSyncer = zapcore.AddSync(os.Stdout)
	} else {
		writeSyncer = zapcore.AddSync(&lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    1024000,
			MaxBackups: 2,
			MaxAge:     48,
			Compress:   true,
			LocalTime:  true,
		})
	}

	// JSON 字段与原 logrus 默认输出完全对齐，日志采集系统无感知
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.LevelKey = "level"
	encoderConfig.MessageKey = "msg"
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// 仅输出 Info 及以上级别，对应原 logrus.InfoLevel
	core := zapcore.NewCore(encoder, writeSyncer, zap.InfoLevel)

	// SugaredLogger 兼容原 logrus 调用风格：Info/Error/Infof/Errorf 等方法签名完全一致
	return zap.New(core).Sugar()
}

func initRpcLog(conf *KitexConf) {
	klog.SetLevel(logLevel(conf))
	klog.SetOutput(os.Stdout)
}

// GetLogger 安全地获取日志实例，延迟初始化逻辑保持不变
func GetLogger() *zap.SugaredLogger {
	return getLog()
}

func logLevel(conf *KitexConf) klog.Level {
	level := conf.Kitex.Server.LogLevel
	switch level {
	case "trace":
		return klog.LevelTrace
	case "debug":
		return klog.LevelDebug
	case "info":
		return klog.LevelInfo
	case "notice":
		return klog.LevelNotice
	case "warn":
		return klog.LevelWarn
	case "error":
		return klog.LevelError
	case "fatal":
		return klog.LevelFatal
	default:
		return klog.LevelInfo
	}
}
