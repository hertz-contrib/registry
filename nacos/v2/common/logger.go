package common

import (
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/nacos-group/nacos-sdk-go/v2/common/logger"
)

type customNacosLogger struct{}

func NewCustomNacosLogger() logger.Logger {
	return customNacosLogger{}
}

func (c customNacosLogger) Info(args ...interface{}) {
	hlog.Info(args)
}

func (c customNacosLogger) Warn(args ...interface{}) {
	hlog.Warn(args)
}

func (c customNacosLogger) Error(args ...interface{}) {
	hlog.Error(args)
}

func (c customNacosLogger) Debug(args ...interface{}) {
	hlog.Debug(args)
}

func (c customNacosLogger) Infof(fmt string, args ...interface{}) {
	hlog.Infof(fmt, args)
}

func (c customNacosLogger) Warnf(fmt string, args ...interface{}) {
	hlog.Warnf(fmt, args)
}

func (c customNacosLogger) Errorf(fmt string, args ...interface{}) {
	hlog.Errorf(fmt, args)
}

func (c customNacosLogger) Debugf(fmt string, args ...interface{}) {
	hlog.Debug(fmt, args)
}
