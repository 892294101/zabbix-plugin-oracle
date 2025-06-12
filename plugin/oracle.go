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
	"context"
	"github.com/892294101/zabbix-agent2-oracle/plugin/handlers"
	"golang.zabbix.com/sdk/metric"
	"golang.zabbix.com/sdk/plugin"
	"golang.zabbix.com/sdk/zbxerr"
	"time"
)

const (
	Name       = "Oracle"
	hkInterval = 10
)

// Plugin -
type Plugin struct {
	plugin.Base
	connMgr *ConnManager
	options PluginOptions
}

var Impl Plugin

func (p *Plugin) Export(key string, rawParams []string, pluginCtx plugin.ContextProvider) (result interface{}, err error) {
	params, _, hc, err := metrics[key].EvalParams(rawParams, p.options.Sessions)
	if err != nil {
		return nil, err
	}
	
	err = metric.SetDefaults(params, hc, p.options.Default)
	if err != nil {
		return nil, err
	}

	handleMetric := getHandlerFunc(key)
	if handleMetric == nil {
		return nil, zbxerr.ErrorUnsupportedMetric
	}

	// 获取连接
	// 连接管理器负责创建和管理连接，确保每个连接在需要时可用。
	// 连接管理器还负责定期检查连接的状态，并在必要时关闭未使用的连接。
	conn, err := p.connMgr.GetConnection(params)
	if err != nil {
		// 如果请求mongodb.ping，则应使用处理连接错误的特殊逻辑，因为如果发生任何错误，它必须返回pingFailed
		if key == keyPing {
			p.Debugf(err.Error())
			return handlers.PingFailed, nil
		}
		p.Errf(err.Error())
		return nil, err
	}

	p.Debugf("Params: %v", params)

	timeout := conn.getTimeout()

	if timeout < time.Second*time.Duration(pluginCtx.Timeout()) {
		timeout = time.Second * time.Duration(pluginCtx.Timeout())
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	result, err = handleMetric(ctx, conn, params)
	if err != nil {
		p.Errf(err.Error())
	}

	return result, err
}

// Start 实现Runner接口，并在插件激活时执行初始化。
func (p *Plugin) Start() {
	handlers.Logger = p.Logger
	p.connMgr = NewConnManager(
		time.Duration(p.options.KeepAlive)*time.Second,
		time.Duration(p.options.Timeout)*time.Second,
		hkInterval*time.Second,
	)
}

// Stop 实现Runner接口，并在插件停用时释放资源。
func (p *Plugin) Stop() {
	p.connMgr = nil
}
