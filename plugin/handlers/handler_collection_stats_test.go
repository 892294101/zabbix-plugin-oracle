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
	"io/ioutil"
	"log"
	"reflect"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"golang.zabbix.com/sdk/zbxerr"
)

func Test_collectionStatsHandler(t *testing.T) {
	var testData map[string]interface{}

	jsonData, err := ioutil.ReadFile("testdata/collStats.json")
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(jsonData, &testData)
	if err != nil {
		log.Fatal(err)
	}

	mockSession := NewMockConn()
	db := mockSession.DB("MyDatabase")
	db.(*MockMongoDatabase).RunFunc = func(dbName, cmd string) ([]byte, error) {
		if cmd == "collStats" {
			return bson.Marshal(testData)
		}

		return nil, errors.New("no such cmd: " + cmd)
	}

	type args struct {
		s      Session
		params map[string]string
	}

	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr error
	}{
		{
			name: "Must parse an output of \" + collStats + \"command",
			args: args{
				s:      mockSession,
				params: map[string]string{"Database": "MyDatabase", "Collection": "MyCollection"},
			},
			want:    strings.TrimSpace(string(jsonData)),
			wantErr: nil,
		},
		{
			name: "Must catch DB.Run() error",
			args: args{
				s:      mockSession,
				params: map[string]string{"Database": mustFail},
			},
			want:    nil,
			wantErr: zbxerr.ErrorCannotFetchData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CollectionStatsHandler(tt.args.s, tt.args.params)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("collectionStatsHandler() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("collectionStatsHandler() got = %v, want %v", got, tt.want)
			}
		})
	}
}
