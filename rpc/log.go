package rpc

import (
	"github.com/cloudwego/kitex/pkg/klog"
	config2 "github.com/flyerxp/lib/config"
	"github.com/flyerxp/lib/logger"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var L *logrus.Logger
var once sync.Once

func init() {
	go getLog()
}
func getLog() *logrus.Logger {
	once.Do(func() {
		paths := logger.GetPath(config2.GetConf().App.Logger.OutputPaths, "rpc")
		logPath := ""
		for _, v := range paths {
			if strings.Contains(v, "/") {
				logPath = filepath.Dir(v) + "/rpc_conn.log"
			}
		}
		pLog := &logrus.Logger{
			Formatter:    new(logrus.JSONFormatter),
			Hooks:        make(logrus.LevelHooks),
			Level:        logrus.InfoLevel,
			ExitFunc:     os.Exit,
			ReportCaller: false,
		}
		if logPath == "" {
			pLog.Out = os.Stdout
		} else {
			pLog.Out = &lumberjack.Logger{
				Filename:   logPath,
				MaxSize:    1024000,
				MaxBackups: 2,
				MaxAge:     48,
				Compress:   true,
				LocalTime:  true,
			}
		}
		L = pLog
	})
	return L
}
func initRpcLog() {
	klog.SetLevel(logLevel())
	klog.SetOutput(os.Stdout)
}
func logLevel() klog.Level {
	level := GetConf().Kitex.Server.LogLevel
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
