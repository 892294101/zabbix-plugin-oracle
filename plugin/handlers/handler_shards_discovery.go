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
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.zabbix.com/sdk/zbxerr"
)

type lldShEntity struct {
	ID        string `json:"{#ID}"`
	Hostname  string `json:"{#HOSTNAME}"`
	MongodURI string `json:"{#MONGOD_URI}"`
	State     string `json:"{#STATE}"`
}

type shEntry struct {
	ID    string      `bson:"_id"`
	Host  string      `bson:"host"`
	State json.Number `bson:"state"`
}

// ShardsDiscoveryHandler
// https://docs.mongodb.com/manual/reference/method/sh.status/#sh.status
func ShardsDiscoveryHandler(s Session, _ map[string]string) (interface{}, error) {
	var shards []shEntry

	opts := options.Find()
	opts.SetSort(bson.D{{Key: sortNatural, Value: 1}})
	opts.SetMaxTime(time.Duration(s.GetMaxTimeMS()) * time.Millisecond)

	q, err := s.DB("config").C("shards").Find(bson.M{}, opts)
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	err = q.Get(&shards)
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	lld := make([]lldShEntity, 0)

	for _, sh := range shards {
		lld, err = handlerShards(sh, lld)
		if err != nil {
			return nil, zbxerr.ErrorCannotParseResult.Wrap(err)
		}
	}

	jsonLLD, err := json.Marshal(lld)
	if err != nil {
		return nil, zbxerr.ErrorCannotMarshalJSON.Wrap(err)
	}

	return string(jsonLLD), nil
}

func handlerShards(sh shEntry, lld []lldShEntity) ([]lldShEntity, error) {
	hosts := sh.Host

	h := strings.SplitN(sh.Host, "/", splitCount)
	if len(h) > 1 {
		hosts = h[1]
	}

	for _, hostport := range strings.Split(hosts, ",") {
		host, _, err := net.SplitHostPort(hostport)
		if err != nil {
			return nil, err
		}

		lld = append(lld, lldShEntity{
			ID:        sh.ID,
			Hostname:  host,
			MongodURI: fmt.Sprintf("%s://%s", UriDefaults.Scheme, hostport),
			State:     sh.State.String(),
		})
	}

	return lld, nil
}
