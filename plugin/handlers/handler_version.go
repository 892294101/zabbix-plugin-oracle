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
	"go.mongodb.org/mongo-driver/bson"
	"golang.zabbix.com/sdk/zbxerr"
)

// VersionHandler executes 'buildInfo' command extracting and returning version
// info from the response.
func VersionHandler(s Session, _ map[string]string) (any, error) {
	buildInfo := bson.M{}

	err := s.DB("admin").Run(&bson.D{{Key: "buildInfo", Value: 1}}, &buildInfo)
	if err != nil {
		return nil, zbxerr.New("failed to run buildInfo command").Wrap(err)
	}

	version, ok := buildInfo["version"]
	if !ok {
		return nil, zbxerr.New("version not found in buildInfo")
	}

	return version, nil
}
