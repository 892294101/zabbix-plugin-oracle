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
)

// handlerFunc defines an interface must be implemented by handlers.
type handlerFunc func(ctx context.Context, s handlers.Database, params map[string]string) (res interface{}, err error)

var metricHandlers = map[string]handlerFunc{
	keyTablespacesUsage: handlers.TablespacesUsageHandler,
	keyPing:             handlers.PingHandler,
}

// getHandlerFunc returns a handlerFunc related to a given key.
func getHandlerFunc(key string) handlerFunc {
	return metricHandlers[key]
}

const (
	keyTablespacesUsage = "oracle.tablespaces.usage"
	keyPing             = "oracle.ping"
)

var (
	paramURI = metric.NewConnParam("URI", "URI to connect or session name.")
)

var metrics = metric.MetricSet{
	keyTablespacesUsage: metric.New("Returns usage statistics for tablespaces.", []*metric.Param{paramURI}, false),
	keyPing:             metric.New("Test if connection is alive or not.", []*metric.Param{paramURI}, false),
}

func init() {
	plugin.RegisterMetrics(&Impl, Name, metrics.List()...)
}
