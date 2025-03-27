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
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.zabbix.com/sdk/zbxerr"
)

const (
	statePrimary   = 1
	stateSecondary = 2
)

const nodeHealthy = 1

type Member struct {
	health int
	optime int
	state  int
	name   string
	ptr    interface{}
}

type rawMember = map[string]interface{}

var errUnknownStructure = errors.New("failed to parse the members structure")

func parseMembers(raw []interface{}) (result []Member, err error) {
	var (
		members     []Member
		primaryNode Member
	)

	for _, m := range raw {
		var member Member
		member, err = paseMember(m)
		if err != nil {
			return
		}

		if member.state == statePrimary {
			primaryNode = member
		} else {
			members = append(members, member)
		}
	}

	result = append([]Member{primaryNode}, members...)
	if len(result) == 0 {
		return nil, errUnknownStructure
	}

	return result, nil
}

func paseMember(m interface{}) (member Member, err error) {
	ok := true

	if v, ok := m.(rawMember)["name"].(string); ok {
		member.name = v
	}

	if v, ok := m.(rawMember)["health"].(float64); ok {
		member.health = int(v)
	}

	if v, ok := m.(rawMember)["optime"].(map[string]interface{}); ok {
		if pa, ok := v["ts"].(primitive.Timestamp); ok {
			member.optime = int(time.Unix(int64(pa.T), 0).Unix())
		} else {
			member.optime = int(int64(v["ts"].(float64)))
		}
	}

	if v, ok := m.(rawMember)["state"].(int32); ok {
		member.state = int(v)
	}

	if !ok {
		return member, errUnknownStructure
	}

	member.ptr = m

	return
}

func injectExtendedMembersStats(raw []interface{}) error {
	members, err := parseMembers(raw)
	if err != nil {
		return err
	}

	unhealthyNodes := []string{}
	unhealthyCount := 0
	primary := members[0]

	for _, node := range members {
		if ptr, ok := node.ptr.(rawMember); ok {
			ptr["lag"] = primary.optime - node.optime
			node.ptr = ptr
		}

		if node.state == stateSecondary && node.health != nodeHealthy {
			unhealthyNodes = append(unhealthyNodes, node.name)
			unhealthyCount++
		}
	}

	if ptr, ok := primary.ptr.(rawMember); ok {
		ptr["unhealthyNodes"] = unhealthyNodes
		ptr["unhealthyCount"] = unhealthyCount
		ptr["totalNodes"] = len(members) - 1
		primary.ptr = ptr
	}

	return nil
}

// ReplSetStatusHandler
// https://docs.mongodb.com/manual/reference/command/replSetGetStatus/index.html
func ReplSetStatusHandler(s Session, _ map[string]string) (interface{}, error) {
	var replSetGetStatus map[string]interface{}

	err := s.DB("admin").Run(
		&bson.D{
			{
				Key:   "replSetGetStatus",
				Value: 1,
			},
			{
				Key:   "maxTimeMS",
				Value: s.GetMaxTimeMS(),
			},
		},
		&replSetGetStatus,
	)

	if err != nil {
		if strings.Contains(err.Error(), "not running with --replSet") {
			return "{}", nil
		}

		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	if pa, ok := replSetGetStatus["members"].(primitive.A); ok {
		Logger.Debugf("members got as primitive A")
		i := []interface{}(pa)
		Logger.Debugf("value:%v\n type: %T\n", i, i)

		err = injectExtendedMembersStats(i)
		if err != nil {
			return nil, zbxerr.ErrorCannotParseResult.Wrap(err)
		}
	}

	jsonRes, err := json.Marshal(replSetGetStatus)
	if err != nil {
		return nil, zbxerr.ErrorCannotMarshalJSON.Wrap(err)
	}

	return string(jsonRes), nil
}
