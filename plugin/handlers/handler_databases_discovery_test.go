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

func Test_databasesDiscoveryHandler(t *testing.T) {
	type args struct {
		s   Session
		dbs []string
	}

	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr error
	}{
		{
			name: "Must return a list of databases",
			args: args{
				s:   NewMockConn(),
				dbs: []string{"testdb", "local", "config"},
			},
			want:    "[{\"{#DBNAME}\":\"config\"},{\"{#DBNAME}\":\"local\"},{\"{#DBNAME}\":\"testdb\"}]",
			wantErr: nil,
		},
		{
			name: "Must catch DB.DatabaseNames() error",
			args: args{
				s:   NewMockConn(),
				dbs: []string{mustFail},
			},
			want:    nil,
			wantErr: zbxerr.ErrorCannotFetchData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, db := range tt.args.dbs {
				tt.args.s.DB(db)
			}

			got, err := DatabasesDiscoveryHandler(tt.args.s, nil)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("databasesDiscoveryHandler() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("databasesDiscoveryHandler() got = %v, want %v", got, tt.want)
			}
		})
	}
}
