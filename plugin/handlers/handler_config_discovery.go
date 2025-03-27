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

package handlers

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"golang.zabbix.com/sdk/zbxerr"
)

type lldCfgEntity struct {
	ReplicaSet string `json:"{#REPLICASET}"`
	Hostname   string `json:"{#HOSTNAME}"`
	MongodURI  string `json:"{#MONGOD_URI}"`
}

type shardMap struct {
	Map map[string]string
}

// ConfigDiscoveryHandler
// https://docs.mongodb.com/manual/reference/command/getShardMap/#dbcmd.getShardMap
func ConfigDiscoveryHandler(s Session, params map[string]string) (interface{}, error) {
	var cfgServers shardMap
	err := s.DB("admin").Run(
		&bson.D{
			{Key: "getShardMap", Value: 1},
			{Key: "maxTimeMS", Value: s.GetMaxTimeMS()},
		},
		&cfgServers,
	)

	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	lld := make([]lldCfgEntity, 0)

	if servers, ok := cfgServers.Map["config"]; ok {
		lld, err = handlerServer(servers, lld)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, zbxerr.ErrorCannotParseResult
	}

	jsonRes, err := json.Marshal(lld)
	if err != nil {
		return nil, zbxerr.ErrorCannotMarshalJSON.Wrap(err)
	}

	return string(jsonRes), nil
}

func handlerServer(servers string, lld []lldCfgEntity) ([]lldCfgEntity, error) {
	var rs string

	hosts := servers

	h := strings.SplitN(hosts, "/", splitCount)
	if len(h) > 1 {
		rs = h[0]
		hosts = h[1]
	}

	for _, hostport := range strings.Split(hosts, ",") {
		host, _, err := net.SplitHostPort(hostport)
		if err != nil {
			return nil, zbxerr.ErrorCannotParseResult.Wrap(err)
		}

		lld = append(lld, lldCfgEntity{
			Hostname:   host,
			MongodURI:  fmt.Sprintf("%s://%s", UriDefaults.Scheme, hostport),
			ReplicaSet: rs,
		})
	}

	return lld, nil
}
