// Copyright 2023 CloudWeGo Authors.
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
	"github.com/nacos-group/nacos-sdk-go/common/logger"
)

type customNacosV2Logger struct{}

func NewCustomNacosV2Logger() logger.Logger {
	return customNacosV2Logger{}
}

func (c customNacosV2Logger) Info(args ...interface{}) {
	hlog.Info(args)
}

func (c customNacosV2Logger) Warn(args ...interface{}) {
	hlog.Warn(args)
}

func (c customNacosV2Logger) Error(args ...interface{}) {
	hlog.Error(args)
}

func (c customNacosV2Logger) Debug(args ...interface{}) {
	hlog.Debug(args)
}

func (c customNacosV2Logger) Infof(fmt string, args ...interface{}) {
	hlog.Infof(fmt, args)
}

func (c customNacosV2Logger) Warnf(fmt string, args ...interface{}) {
	hlog.Warnf(fmt, args)
}

func (c customNacosV2Logger) Errorf(fmt string, args ...interface{}) {
	hlog.Errorf(fmt, args)
}

func (c customNacosV2Logger) Debugf(fmt string, args ...interface{}) {
	hlog.Debug(fmt, args)
}
