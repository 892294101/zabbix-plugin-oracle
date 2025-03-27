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
)

func Test_collectionsUsageHandler(t *testing.T) {
	var testData map[string]interface{}

	jsonData, err := ioutil.ReadFile("testdata/top.json")
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(jsonData, &testData)
	if err != nil {
		log.Fatal(err)
	}

	mockSession := NewMockConn()
	db := mockSession.DB("admin")
	db.(*MockMongoDatabase).RunFunc = func(dbName, cmd string) ([]byte, error) {
		if cmd == "top" {
			return bson.Marshal(testData)
		}

		return nil, errors.New("no such cmd: " + cmd)
	}

	type args struct {
		s Session
	}

	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr error
	}{
		{
			name: "Must parse an output of \" + top + \"command",
			args: args{
				s: mockSession,
			},
			want:    strings.TrimSpace(string(jsonData)),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CollectionsUsageHandler(tt.args.s, nil)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("collectionsUsageHandler() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("collectionsUsageHandler() got = %v, want %v", got, tt.want)
			}
		})
	}
}
