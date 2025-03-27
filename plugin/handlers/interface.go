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

	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.zabbix.com/sdk/log"
	"golang.zabbix.com/sdk/uri"
)

const (
	PingFailed = 0
	PingOk     = 1

	splitCount = 2

	sortNatural = "$natural"

	timestampBSONName = "ts"
)

var UriDefaults = &uri.Defaults{Scheme: "tcp", Port: "27017"}

var Logger log.Logger

var errNotFound = errors.New("not found")

// Session is an interface to access to the session struct.
type Session interface {
	DB(name string) Database
	DatabaseNames() (names []string, err error)
	GetMaxTimeMS() int64
	Ping() error
}

type Database interface {
	C(name string) Collection
	CollectionNames() (names []string, err error)
	Run(cmd, result interface{}) error
}

type Collection interface {
	Find(query interface{}, opts ...*options.FindOptions) (q Query, err error)
	FindOne(query interface{}, opts ...*options.FindOneOptions) Query
}

type Query interface {
	Count() (n int, err error)
	Get(result interface{}) error
	GetSingle(result interface{}) error
}
