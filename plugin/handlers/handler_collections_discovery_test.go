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
	"errors"
	"reflect"
	"testing"

	"golang.zabbix.com/sdk/zbxerr"
)

func Test_collectionsDiscoveryHandler(t *testing.T) {
	type args struct {
		s   Session
		dbs map[string][]string
	}

	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr error
	}{
		{
			name: "Must return a list of collections",
			args: args{
				s: NewMockConn(),
				dbs: map[string][]string{
					"testdb": {"col1", "col2"},
					"local":  {"startup_log"},
					"config": {"system.sessions"},
				},
			},
			want: "[{\"{#COLLECTION}\":\"system.sessions\",\"{#DBNAME}\":\"config\"},{\"{#COLLECTION}\":" +
				"\"startup_log\",\"{#DBNAME}\":\"local\"},{\"{#COLLECTION}\":\"col1\",\"{#DBNAME}\":\"testdb\"}," +
				"{\"{#COLLECTION}\":\"col2\",\"{#DBNAME}\":\"testdb\"}]",
			wantErr: nil,
		},
		{
			name: "Must catch DB.DatabaseNames() error",
			args: args{
				s:   NewMockConn(),
				dbs: map[string][]string{mustFail: {}},
			},
			want:    nil,
			wantErr: zbxerr.ErrorCannotFetchData,
		},
		{
			name: "Must catch DB.CollectionNames() error",
			args: args{
				s:   NewMockConn(),
				dbs: map[string][]string{"MyDatabase": {mustFail}},
			},
			want:    nil,
			wantErr: zbxerr.ErrorCannotFetchData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for db, cc := range tt.args.dbs {
				tt.args.s.DB(db)
				for _, c := range cc {
					tt.args.s.DB(db).C(c)
				}
			}

			got, err := CollectionsDiscoveryHandler(tt.args.s, nil)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("collectionsDiscoveryHandler() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("collectionsDiscoveryHandler() got = %v, want %v", got, tt.want)
			}
		})
	}
}
