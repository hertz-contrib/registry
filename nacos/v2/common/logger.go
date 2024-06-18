// Copyright 2021 CloudWeGo Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"github.com/cloudwego/hertz/pkg/common/hlog"
	v2 "github.com/nacos-group/nacos-sdk-go/v2/common/logger"
)

type customNacosLogger struct{}

func NewCustomNacosLogger() v2.Logger {
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
