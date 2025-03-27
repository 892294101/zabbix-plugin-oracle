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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.zabbix.com/sdk/zbxerr"
)

const (
	mustFail      = "mustFail"
	mockTimestamp = 3000
)

type MockConn struct {
	dbs map[string]*MockMongoDatabase
}

func NewMockConn() *MockConn {
	return &MockConn{
		dbs: make(map[string]*MockMongoDatabase),
	}
}

func (conn *MockConn) DB(name string) Database {
	if db, ok := conn.dbs[name]; ok {
		return db
	}

	conn.dbs[name] = &MockMongoDatabase{
		name:        name,
		collections: make(map[string]*MockMongoCollection),
	}

	return conn.dbs[name]
}

func (conn *MockConn) DatabaseNames() (names []string, err error) {
	for _, db := range conn.dbs {
		if db.name == mustFail {
			return nil, zbxerr.ErrorCannotFetchData
		}

		names = append(names, db.name)
	}

	return
}

func (conn *MockConn) Ping() error {
	return nil
}

func (conn *MockConn) GetMaxTimeMS() int64 {
	return mockTimestamp
}

type MockSession interface {
	DB(name string) Database
	DatabaseNames() (names []string, err error)
	GetMaxTimeMS() int64
	Ping() error
}

type MockMongoDatabase struct {
	name        string
	collections map[string]*MockMongoCollection
	RunFunc     func(dbName, cmd string) ([]byte, error)
}

func (d *MockMongoDatabase) C(name string) Collection {
	if col, ok := d.collections[name]; ok {
		return col
	}

	d.collections[name] = &MockMongoCollection{
		name:    name,
		queries: make(map[interface{}]*MockMongoQuery),
	}

	return d.collections[name]
}

func (d *MockMongoDatabase) CollectionNames() (names []string, err error) {
	for _, col := range d.collections {
		if col.name == mustFail {
			return nil, errors.New("fail")
		}

		names = append(names, col.name)
	}

	return
}

func (d *MockMongoDatabase) Run(cmd, result interface{}) error {
	if d.RunFunc == nil {
		d.RunFunc = func(dbName, _ string) ([]byte, error) {
			if dbName == mustFail {
				return nil, errors.New("fail")
			}

			return bson.Marshal(map[string]int{"ok": 1})
		}
	}

	if result == nil {
		return nil
	}

	bsonDcmd := *(cmd.(*bson.D))
	cmdName := bsonDcmd[0].Key

	data, err := d.RunFunc(d.name, cmdName)
	if err != nil {
		return err
	}

	return bson.Unmarshal(data, result)
}

type MockMongoCollection struct {
	name    string
	queries map[interface{}]*MockMongoQuery
}

func (c *MockMongoCollection) Find(query interface{}, opts ...*options.FindOptions) (q Query, err error) {
	queryHash := fmt.Sprintf("%v", query)
	if q, ok := c.queries[queryHash]; ok {
		return q, nil
	}

	c.queries[queryHash] = &MockMongoQuery{
		collection: c.name,
		query:      query,
	}

	return c.queries[queryHash], nil
}

func (c *MockMongoCollection) FindOne(query interface{}, opts ...*options.FindOneOptions) Query {
	queryHash := fmt.Sprintf("%v", query)
	if q, ok := c.queries[queryHash]; ok {
		return q
	}

	c.queries[queryHash] = &MockMongoQuery{
		collection: c.name,
		query:      query,
	}

	return c.queries[queryHash]
}

type MockMongoQuery struct {
	collection string
	query      interface{}
	DataFunc   func() ([]byte, error)
}

func (q *MockMongoQuery) retrieve(result interface{}) error {
	if q.DataFunc == nil {
		return errNotFound
	}

	if result == nil {
		return nil
	}

	data, err := q.DataFunc()
	if err != nil {
		return err
	}

	return bson.Unmarshal(data, result)
}

func (q *MockMongoQuery) Count() (n int, err error) {
	return 1, nil
}

func (q *MockMongoQuery) Get(result interface{}) error {
	return q.retrieve(result)
}

func (q *MockMongoQuery) GetSingle(result interface{}) error {
	return q.retrieve(result)
}
