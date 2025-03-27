/*
** Zabbix
** Copyright (C) 2001-2025 Zabbix SIA
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
	"time"

	"golang.zabbix.com/plugin/mongodb/plugin/handlers"
	"golang.zabbix.com/sdk/metric"
	"golang.zabbix.com/sdk/plugin"
	"golang.zabbix.com/sdk/uri"
	"golang.zabbix.com/sdk/zbxerr"
)

const (
	Name       = "MongoDB"
	hkInterval = 10
)

var Impl Plugin

// Plugin -
type Plugin struct {
	plugin.Base
	connMgr *ConnManager
	options PluginOptions
}

func (p *Plugin) Export(key string, rawParams []string, ctx plugin.ContextProvider) (result interface{}, err error) {
	params, _, hc, err := metrics[key].EvalParams(rawParams, p.options.Sessions)
	if err != nil {
		return nil, err
	}

	err = metric.SetDefaults(params, hc, p.options.Default)
	if err != nil {
		return nil, err
	}

	uri, err := uri.NewWithCreds(params["URI"], params["User"], params["Password"], handlers.UriDefaults)
	if err != nil {
		return nil, err
	}

	handleMetric := getHandlerFunc(key)
	if handleMetric == nil {
		return nil, zbxerr.ErrorUnsupportedMetric
	}

	conn, err := p.connMgr.GetConnection(*uri, params)
	if err != nil {
		// Special logic of processing connection errors should be used if mongodb.ping is requested
		// because it must return pingFailed if any error occurred.
		if key == keyPing {
			p.Debugf(err.Error())

			return handlers.PingFailed, nil
		}

		p.Errf(err.Error())

		return nil, err
	}

	p.Debugf("Params: %v", params)

	result, err = handleMetric(conn, params)
	if err != nil {
		p.Errf(err.Error())
	}

	return result, err
}

// Start implements the Runner interface and performs initialization when plugin is activated.
func (p *Plugin) Start() {
	handlers.Logger = p.Logger
	p.connMgr = NewConnManager(
		time.Duration(p.options.KeepAlive)*time.Second,
		time.Duration(p.options.Timeout)*time.Second,
		hkInterval*time.Second,
		p.Logger,
	)
}

// Stop implements the Runner interface and frees resources when plugin is deactivated.
func (p *Plugin) Stop() {
	p.connMgr.Destroy()
	p.connMgr = nil
}
