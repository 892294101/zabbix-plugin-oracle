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

// PingHandler executes 'ping' command and returns pingOk if a connection is alive or pingFailed otherwise.
// https://docs.mongodb.com/manual/reference/command/ping/index.html
func PingHandler(s Session, _ map[string]string) (interface{}, error) {
	if err := s.Ping(); err != nil {
		Logger.Debugf("ping failed, %s", err.Error())

		return PingFailed, nil
	}

	return PingOk, nil
}
