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
	"sort"

	"golang.zabbix.com/sdk/zbxerr"
)

type colEntity struct {
	ColName string `json:"{#COLLECTION}"`
	DbName  string `json:"{#DBNAME}"`
}

// CollectionsDiscoveryHandler
// https://docs.mongodb.com/manual/reference/command/listDatabases/
func CollectionsDiscoveryHandler(s Session, _ map[string]string) (interface{}, error) {
	dbs, err := s.DatabaseNames()
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	sort.Strings(dbs)

	lld := make([]colEntity, 0)

	for _, db := range dbs {
		collections, err := s.DB(db).CollectionNames()

		sort.Strings(collections)

		if err != nil {
			return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
		}

		for _, col := range collections {
			lld = append(lld, colEntity{
				ColName: col,
				DbName:  db,
			})
		}
	}

	jsonLLD, err := json.Marshal(lld)
	if err != nil {
		return nil, zbxerr.ErrorCannotMarshalJSON.Wrap(err)
	}

	return string(jsonLLD), nil
}
