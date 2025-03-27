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
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func Test_oplogStatsHandler(t *testing.T) {
	t.Parallel()

	var (
		opFirst     = &bson.D{{Key: "ts", Value: &primitive.Timestamp{T: uint32(6644097), I: 1}}}
		opLast      = &bson.D{{Key: "ts", Value: &primitive.Timestamp{T: uint32(2178177), I: 1}}}
		oplogQuery  = fmt.Sprintf("%v", bson.M{"ts": bson.M{"$exists": true}})
		newDataFunc = func(
			resps []any, errs []error,
		) func() ([]byte, error) {
			var counter int

			return func() ([]byte, error) {
				defer func() { counter++ }()

				if errs[counter] != nil {
					return nil, errs[counter]
				}

				return bson.Marshal(resps[counter])
			}
		}
	)

	type fields struct {
		collections map[string]*MockMongoCollection
	}

	tests := []struct {
		name    string
		fields  fields
		want    interface{}
		wantErr bool
	}{
		{
			"+ oplog.rs collection",
			fields{
				collections: map[string]*MockMongoCollection{
					"oplog.rs": {
						queries: map[any]*MockMongoQuery{
							oplogQuery: {
								DataFunc: newDataFunc(
									[]any{opFirst, opLast},
									[]error{nil, nil},
								),
							},
						},
					},
				},
			},
			`{"timediff":4465920}`,
			false,
		},
		{
			"+ oplog.$main collection",
			fields{
				collections: map[string]*MockMongoCollection{
					"oplog.rs": {
						queries: map[any]*MockMongoQuery{
							oplogQuery: {
								DataFunc: newDataFunc(
									[]any{nil},
									[]error{mongo.ErrNoDocuments},
								),
							},
						},
					},
					"oplog.$main": {
						queries: map[any]*MockMongoQuery{
							oplogQuery: {
								DataFunc: newDataFunc(
									[]any{opFirst, opLast},
									[]error{nil, nil},
								),
							},
						},
					},
				},
			},
			`{"timediff":4465920}`,
			false,
		},
		{
			"-getTSError",
			fields{
				collections: map[string]*MockMongoCollection{
					"oplog.rs": {
						queries: map[any]*MockMongoQuery{
							oplogQuery: {
								DataFunc: newDataFunc(
									[]any{opFirst},
									[]error{errors.New("fail")},
								),
							},
						},
					},
				},
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
				map[string]*MockMongoDatabase{
					"local": {collections: tt.fields.collections},
				},
			}

			got, err := OplogStatsHandler(mockSess, nil)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"oplogStatsHandler() error = %v, wantErr %v",
					err, tt.wantErr,
				)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("oplogStatsHandler() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
