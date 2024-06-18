package main

import (
	"github.com/hertz-contrib/registry/nacos/v2/common"
	"github.com/nacos-group/nacos-sdk-go/v2/common/logger"
)

func main() {
	logger.SetLogger(common.NewCustomNacosLogger())
	logger.Info("Hello, Nacos!")
}
