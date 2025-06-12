/*
** Zabbix
** Copyright 2001-2022 Zabbix SIA
**
** Licensed under the Apache License, Version 2.0 (the "License");
** you may not use this file except in compliance with the License.
** You may obtain a copy of the License at
**
**     http://www.apache.org/licenses/LICENSE-2.0
**
** Unless required by applicable law or agreed to in writing, software
** distributed under the License is distributed on an "AS IS" BASIS,
** WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
** See the License for the specific language governing permissions and
** limitations under the License.
**/

package plugin

import (
	"fmt"
	"golang.zabbix.com/sdk/conf"
	"golang.zabbix.com/sdk/plugin"
)

type Session struct {
	URI        string `conf:"name=Uri"`                                // 连接字符串
	MinIdle    string `conf:"name=MinIdle,range=1:100,default=5"`      // 最小空闲连接数
	MaxConnect string `conf:"name=MaxConnect,range=1:200,default=100"` // 最大连接数
}

type PluginOptions struct {
	plugin.SystemOptions `conf:"optional,name=System"`
	// 在会话中首次连接以及后续操作时，等待服务器响应的时间量
	Timeout int `conf:"optional,range=1:30"`

	// KeepAlive  未使用连接关闭前的等待时间
	KeepAlive int `conf:"optional,range=60:900,default=60"`

	// 存储预定义的命名连接设置集合
	// 每个连接都有一个唯一的名称，用于在插件配置中引用
	Sessions map[string]Session `conf:"optional"`
	Default  Session            `conf:"optional"`
}

// Configure 实现配置接口
// 初始化配置结构
func (p *Plugin) Configure(global *plugin.GlobalOptions, options interface{}) {
	if err := conf.Unmarshal(options, &p.options); err != nil {
		p.Errf("cannot unmarshal configuration options: %s", err)
	}

	if p.options.Timeout == 0 {
		p.options.Timeout = global.Timeout
	}
}

// Validate 实现配置接口
// 如果效验插件配置失败将返回错误
func (p *Plugin) Validate(options interface{}) error {
	var opts PluginOptions

	err := conf.Unmarshal(options, &opts)
	if err != nil {
		return err
	}

	for s, session := range opts.Sessions {
		fmt.Println("options.Sessions: ", s, session.URI, session.MinIdle, session.MaxConnect)
	}

	return nil
}
