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

	"go.mongodb.org/mongo-driver/bson"
	"golang.zabbix.com/sdk/zbxerr"
)

// ConnPoolStatsHandler
// https://docs.mongodb.com/manual/reference/command/connPoolStats/#dbcmd.connPoolStats
func ConnPoolStatsHandler(s Session, params map[string]string) (interface{}, error) {
	connPoolStats := &bson.M{}
	err := s.DB("test").Run(
		&bson.D{
			{Key: "connPoolStats", Value: 1},
			{Key: "maxTimeMS", Value: s.GetMaxTimeMS()},
		},
		connPoolStats,
	)
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	jsonRes, err := json.Marshal(connPoolStats)
	if err != nil {
		return nil, zbxerr.ErrorCannotMarshalJSON.Wrap(err)
	}

	return string(jsonRes), nil
}
