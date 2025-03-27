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
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.mongodb.org/mongo-driver/bson"
)

func TestVersionHandler(t *testing.T) {
	t.Parallel()

	sampleResp := bson.M{
		"version":           "4.4.0",
		"debug":             false,
		"gitVersion":        "90c65f9cc8fc4e6664a5848230abaa9b3f3b02f7",
		"javascriptEngine":  "mozjs",
		"maxBsonObjectSize": 16777216,
		"ok":                1,
	}

	type db struct {
		resp bson.M
		err  error
	}

	tests := []struct {
		name    string
		db      db
		want    any
		wantErr bool
	}{
		{
			"+valid",
			db{sampleResp, nil},
			any("4.4.0"),
			false,
		},
		{
			"-commandErr",
			db{sampleResp, errors.New("fail")},
			nil,
			true,
		},
		{
			"-missingVersion",
			db{
				bson.M{
					"debug":             false,
					"gitVersion":        "90c65f9cc8fc4e6664a5848230abaa9b3f3b02f7",
					"javascriptEngine":  "mozjs",
					"maxBsonObjectSize": 16777216,
					"ok":                1,
				},
				nil,
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSess := &MockConn{
				dbs: map[string]*MockMongoDatabase{
					"admin": {
						RunFunc: func(_, _ string) ([]byte, error) {
							if tt.db.err != nil {
								return nil, tt.db.err
							}

							b, err := bson.Marshal(tt.db.resp)
							if err != nil {
								t.Fatalf("failed to marshal response: %v", err)
							}

							return b, nil
						},
					},
				},
			}

			got, err := VersionHandler(mockSess, nil)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"VersionHandler() error = %v, wantErr %v", err, tt.wantErr,
				)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("VersionHandler() = %s", diff)
			}
		})
	}
}
