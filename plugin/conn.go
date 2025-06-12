/*
** Zabbix
** Copyright 2001-2022 Zabbix SIA
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

package plugin

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/892294101/zabbix-agent2-oracle/plugin/handlers"
	_ "github.com/godror/godror"
	"golang.zabbix.com/sdk/uri"
	"sync"
	"time"
)

type OracleConn struct {
	addr           string
	timeout        time.Duration
	lastTimeAccess time.Time
	session        *sql.DB
}

func (conn *OracleConn) Database(name string) handlers.Session {
	//TODO implement me
	panic("implement me")
}

func (conn *OracleConn) Ping(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

type ConnManager struct {
	sync.Mutex
	connMutex   sync.Mutex
	connections map[uri.URI]*OracleConn
	keepAlive   time.Duration
	timeout     time.Duration
}

func (conn *OracleConn) getTimeout() time.Duration {
	return conn.timeout
}

// NewConnManager 初始化connManager结构并运行Go例程，该例程监视未使用的连接。
func NewConnManager(keepAlive, timeout, hkInterval time.Duration) *ConnManager {
	connMgr := &ConnManager{connections: make(map[uri.URI]*OracleConn), keepAlive: keepAlive, timeout: timeout}
	return connMgr
}

func (c *ConnManager) GetConnection(params map[string]string) (*OracleConn, error) {
	fmt.Println("GetConnection: ", params)
	return nil, nil
}
