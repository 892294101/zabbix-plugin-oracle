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
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.zabbix.com/sdk/zbxerr"
)

// OplogStatsHandler
// https://docs.mongodb.com/manual/reference/method/db.getReplicationInfo/index.html
func OplogStatsHandler(s Session, _ map[string]string) (any, error) {
	var (
		err             error
		firstTs, lastTs int
	)

	localDb := s.DB("local")
	findOptions := options.FindOne()
	findOptions.SetMaxTime(time.Duration(s.GetMaxTimeMS()) * time.Millisecond)

	for _, collection := range []string{
		"oplog.rs",    // the capped collection that holds the oplog for Replica Set Members
		"oplog.$main", // oplog for the master-slave configuration
	} {
		firstTs, lastTs, err = getTS(collection, localDb, findOptions)
		if err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
			}

			continue
		}

		break
	}

	jsonRes, err := json.Marshal(
		struct {
			TimeDiff int `json:"timediff"` // in seconds
		}{
			TimeDiff: firstTs - lastTs,
		},
	)
	if err != nil {
		return nil, zbxerr.ErrorCannotMarshalJSON.Wrap(err)
	}

	return string(jsonRes), nil
}

func getTS(
	collection string,
	localDb Database,
	findOptions *options.FindOneOptions,
) (int, int, error) {
	findOptions.SetSort(bson.D{{Key: sortNatural, Value: -1}})

	firstTs, err := getOplogStats(localDb, collection, findOptions)
	if err != nil {
		return 0, 0, err
	}

	findOptions.SetSort(bson.D{{Key: sortNatural, Value: 1}})

	lastTs, err := getOplogStats(localDb, collection, findOptions)
	if err != nil {
		return 0, 0, err
	}

	return firstTs, lastTs, nil
}

func getOplogStats(
	db Database,
	collection string,
	opt *options.FindOneOptions,
) (int, error) {
	var result primitive.D

	err := db.C(collection).FindOne(bson.M{"ts": bson.M{"$exists": true}}, opt).
		GetSingle(&result)
	if err != nil {
		return 0, err
	}

	var out int

	for _, op := range result {
		if op.Key == timestampBSONName {
			if pt, ok := op.Value.(primitive.Timestamp); ok {
				out = int(time.Unix(int64(pt.T), 0).Unix())
			}
		}
	}

	return out, nil
}
