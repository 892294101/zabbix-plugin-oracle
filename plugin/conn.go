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

package plugin

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.zabbix.com/plugin/mongodb/plugin/handlers"
	"golang.zabbix.com/sdk/errs"
	"golang.zabbix.com/sdk/log"
	"golang.zabbix.com/sdk/tlsconfig"
	"golang.zabbix.com/sdk/uri"
	"golang.zabbix.com/sdk/zbxerr"
)

const (
	// connType
	disable    = "not_tls"
	require    = "required"
	verifyCa   = "verify_ca"
	verifyFull = "verify_full"
)

type MongoConn struct {
	addr           string
	timeout        time.Duration
	lastTimeAccess time.Time
	session        mongo.Session
}

// MongoDatabase wraps a mgo.Database to embed methods in models.
type MongoDatabase struct {
	// *mgo.Database
	*mongo.Database
}

// MongoCollection wraps a mongo.Collection to embed methods in models.
type MongoCollection struct {
	*mongo.Collection
}

// Query is an interface to access to the query struct

// MongoQuery wraps a mgo.Query to embed methods in models.
type MongoQuery struct {
	*mongo.Cursor
	*mongo.SingleResult
}

// ConnManager is thread-safe structure for manage connections.
type ConnManager struct {
	connectionsMu sync.Mutex
	connections   map[connKey]*MongoConn
	keepAlive     time.Duration
	timeout       time.Duration
	Destroy       context.CancelFunc
	log           log.Logger
}

type connKey struct {
	uri        uri.URI
	rawUri     string
	tlsConnect string
	tlsCA      string
	tlsCert    string
	tlsKey     string
}

// DB shadows *mgo.DB to returns a Database interface instead of *mgo.Database.
func (conn *MongoConn) DB(name string) handlers.Database {
	conn.checkConnection()

	return &MongoDatabase{Database: conn.session.Client().Database(name)}
}

// DatabaseNames returns a list of database names.
func (conn *MongoConn) DatabaseNames() ([]string, error) {
	conn.checkConnection()

	names, err := conn.session.Client().
		ListDatabaseNames(context.Background(), bson.D{})
	if err != nil {
		return nil, errs.Wrap(err, "failed to list database names")
	}

	return names, nil
}

func (conn *MongoConn) Ping() error {
	Impl.Debugf("executing ping for address: %s", conn.addr)
	result := conn.session.Client().Database("admin").RunCommand(
		context.Background(),
		&bson.D{
			{Key: "ping", Value: 1},
			{Key: "maxTimeMS", Value: conn.GetMaxTimeMS()},
		},
	)

	err := result.Err()
	if err != nil {
		Impl.Debugf("failed to ping database %s, %s", conn.addr, err.Error())

		return err
	}

	Impl.Debugf("ping successful for address: %s", conn.addr)

	return nil
}

func (conn *MongoConn) GetMaxTimeMS() int64 {
	return conn.timeout.Milliseconds()
}

// C shadows *mongo.DB to returns a Database interface instead of *mgo.Database.
func (d *MongoDatabase) C(name string) handlers.Collection {
	return &MongoCollection{Collection: d.Collection(name)}
}

func (d *MongoDatabase) CollectionNames() (names []string, err error) {
	return d.Database.ListCollectionNames(context.Background(), bson.D{})
}

// Run shadows *mgo.DB to returns a Database interface instead of *mgo.Database.
func (d *MongoDatabase) Run(cmd, result interface{}) error {
	return d.Database.RunCommand(context.Background(), cmd).Decode(result)
}

// Collection is an interface to access to the collection struct.

// Find shadows *mgo.Collection to returns a Query interface instead of *mgo.Query.
func (c *MongoCollection) Find( //nolint:ireturn
	query any,
	opts ...*options.FindOptions,
) (handlers.Query, error) {
	cursor, err := c.Collection.Find(context.Background(), query, opts...)
	if err != nil {
		return nil, errs.Wrap(err, "failed to execute find query")
	}

	return &MongoQuery{Cursor: cursor}, nil
}

// FindOne shadows *mgo.Collection to returns a Query interface instead of *mgo.Query.
func (c *MongoCollection) FindOne( //nolint:ireturn
	query any,
	opts ...*options.FindOneOptions,
) handlers.Query {
	return &MongoQuery{
		SingleResult: c.Collection.FindOne(
			context.Background(),
			query,
			opts...),
	}
}

func (q *MongoQuery) Count() (int, error) {
	var in []interface{}
	err := q.Cursor.All(context.Background(), &in)
	if err != nil {
		return 0, err
	}

	return len(in), nil
}

func (q *MongoQuery) Get(result interface{}) error {
	if q.Cursor.Err() != nil {
		return q.Cursor.Err()
	}

	return q.Cursor.All(context.Background(), result)
}

func (q *MongoQuery) GetSingle(result interface{}) error {
	if q.SingleResult.Err() != nil {
		return q.SingleResult.Err()
	}

	return q.SingleResult.Decode(result)
}

// NewConnManager initializes connManager structure and runs Go Routine that watches for unused connections.
func NewConnManager(
	keepAlive, timeout, hkInterval time.Duration,
	logger log.Logger,
) *ConnManager {
	ctx, cancel := context.WithCancel(context.Background())

	connMgr := &ConnManager{
		connections: make(map[connKey]*MongoConn),
		keepAlive:   keepAlive,
		timeout:     timeout,
		Destroy:     cancel, // Destroy stops originated goroutines and close connections.
		log:         logger,
	}

	go connMgr.housekeeper(ctx, hkInterval)

	return connMgr
}

// GetConnection returns an existing connection or creates a new one.
func (c *ConnManager) GetConnection(
	connURI uri.URI, //nolint:gocritic
	params map[string]string,
) (*MongoConn, error) {
	ck := createConnKey(connURI, params)

	conn := c.getConn(ck)
	if conn != nil {
		c.log.Tracef("connection found for host: %s", connURI.Host())

		return conn, nil
	}

	conn, err := c.create(ck, params)
	if err != nil {
		return nil, errs.Wrap(err, "failed to create new connection")
	}

	return c.setConn(ck, conn), nil
}

// getConn returns a connection with given uri if it exists and also updates
// lastTimeAccess, otherwise returns nil.
func (c *ConnManager) getConn(ck connKey) *MongoConn { //nolint:gocritic
	c.connectionsMu.Lock()
	defer c.connectionsMu.Unlock()

	conn, ok := c.connections[ck]
	if !ok {
		return nil
	}

	conn.updateAccessTime()

	return conn
}

func (c *ConnManager) setConn(
	ck connKey, //nolint:gocritic
	conn *MongoConn,
) *MongoConn {
	c.connectionsMu.Lock()
	defer c.connectionsMu.Unlock()

	existingConn, ok := c.connections[ck]
	if ok {
		err := closeSession(context.Background(), conn.session)
		if err != nil {
			c.log.Warningf("set conn session client clean-up failed: %s", err.Error())
		}

		c.log.Debugf("Closed unused connection: %s", ck.uri.Addr())

		return existingConn
	}

	c.connections[ck] = conn

	return conn
}

// closeUnused closes each connection that has not been accessed at least within the keepalive interval.
func (c *ConnManager) closeUnused() {
	c.connectionsMu.Lock()
	defer c.connectionsMu.Unlock()

	for ck, conn := range c.connections {
		if time.Since(conn.lastTimeAccess) > c.keepAlive {
			err := closeSession(context.Background(), conn.session)
			if err != nil {
				c.log.Warningf("unused session client clean-up failed: %s", err.Error())
			}

			delete(c.connections, ck)
			c.log.Debugf("Closed unused connection: %s", ck.uri.Addr())
		}
	}
}

// closeAll closes all existed connections.
func (c *ConnManager) closeAll() {
	c.connectionsMu.Lock()
	for uri, conn := range c.connections {
		err := closeSession(context.Background(), conn.session)
		if err != nil {
			c.log.Warningf("close all session client clean-up failed: %s", err.Error())
		}

		delete(c.connections, uri)
	}
	c.connectionsMu.Unlock()
	c.log.Debugf("Closed all connections")
}

// housekeeper repeatedly checks for unused connections and close them.
func (c *ConnManager) housekeeper(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			c.closeAll()

			return
		case <-ticker.C:
			c.closeUnused()
		}
	}
}

// create creates a new connection with given credentials.
func (c *ConnManager) create(
	ck connKey, //nolint:gocritic
	params map[string]string,
) (*MongoConn, error) {
	opt, err := c.createOptions(ck.uri, params)
	if err != nil {
		return nil, err
	}

	client, err := mongo.NewClient(opt)
	if err != nil {
		return nil, err
	}

	err = client.Connect(context.Background())
	if err != nil {
		return nil, err
	}

	session, err := client.StartSession()
	if err != nil {
		c.log.Debugf("session start failed failed: %s", ck.uri.Addr())

		cerr := closeSession(context.Background(), session)
		if cerr != nil {
			c.log.Warningf("session start clean-up failed: %s", cerr.Error())
		}

		return nil, err
	}

	err = session.Client().Ping(context.Background(), readpref.Nearest())
	if err != nil {
		c.log.Debugf("session client ping failed: %s", ck.uri.Addr())

		cerr := closeSession(context.Background(), session)
		if cerr != nil {
			c.log.Warningf("session client ping clean-up failed: %s", cerr.Error())
		}

		return nil, err
	}

	c.log.Debugf("Created new connection: %s", ck.uri.Addr())

	return &MongoConn{
		addr:           ck.uri.Addr(),
		timeout:        c.timeout,
		lastTimeAccess: time.Now(),
		session:        session,
	}, nil
}

// get returns a connection with given uri if it exists and also updates lastTimeAccess, otherwise returns nil.
func (c *ConnManager) createOptions(
	connURI uri.URI, //nolint:gocritic
	params map[string]string,
) (*options.ClientOptions, error) {
	details, err := createTLS(params)
	if err != nil {
		return nil, err
	}

	opt := options.Client()

	if connURI.User() != "" {
		creds := options.Credential{}
		creds.Username = connURI.User()
		creds.Password = connURI.Password()
		creds.PasswordSet = true
		opt = opt.SetAuth(creds)
	}

	if details.TlsConnect != disable {
		err := c.setTLSConfig(opt, details)
		if err != nil {
			return nil, err
		}
	}

	opt.SetHosts([]string{connURI.Addr()})
	opt.SetDirect(true)
	opt.SetConnectTimeout(c.timeout)
	opt.SetServerSelectionTimeout(c.timeout)
	opt.SetMaxPoolSize(1)

	return opt, nil
}

func createTLS(params map[string]string) (*tlsconfig.Details, error) {
	var (
		validateCA     = true
		validateClient = false
		tlsType        = params[tlsConnectParam]
	)

	if tlsType == "" {
		tlsType = disable
	}

	details := tlsconfig.NewDetails(
		"",
		tlsType,
		params[tlsCAParam],
		params[tlsCertParam],
		params[tlsKeyParam],
		params[uriParam],
		disable,
		require,
		verifyCa,
		verifyFull,
	)

	if tlsType == disable || tlsType == require {
		validateCA = false
	}

	if details.TlsKeyFile != "" || details.TlsCertFile != "" {
		validateClient = true
	}

	err := details.Validate(validateCA, validateClient, validateClient)
	if err != nil {
		return nil, zbxerr.ErrorInvalidConfiguration.Wrap(err)
	}

	return &details, nil
}

func (c *ConnManager) setTLSConfig(
	opt *options.ClientOptions,
	details *tlsconfig.Details,
) error {
	var cfg *tls.Config
	var err error

	switch details.TlsConnect {
	case "required":
		cfg, err = c.getRequiredTLSConfig(details)
		if err != nil {
			return err
		}
	case "verify_ca":
		cfg, err = details.GetTLSConfig(true)
		if err != nil {
			return errs.Wrap(err, "failed to get TLS config for verify_ca connection")
		}

		cfg.VerifyPeerCertificate = tlsconfig.VerifyPeerCertificateFunc(
			"",
			cfg.RootCAs,
		)
	case "verify_full":
		cfg, err = details.GetTLSConfig(false)
		if err != nil {
			return errs.Wrap(err, "failed to get TLS config for verify_full connection")
		}
	}

	opt.SetTLSConfig(cfg)

	return nil
}

func (c *ConnManager) getRequiredTLSConfig(
	details *tlsconfig.Details,
) (*tls.Config, error) {
	if details.TlsCaFile != "" {
		c.log.Warningf(
			"server CA will not be verified for %s",
			details.TlsConnect,
		)
	}

	clientCerts, err := details.LoadCertificates()
	if err != nil {
		return nil, err
	}

	return &tls.Config{Certificates: clientCerts, InsecureSkipVerify: true}, nil
}

// updateAccessTime updates the last time a connection was accessed.
func (conn *MongoConn) updateAccessTime() {
	conn.lastTimeAccess = time.Now()
}

// checkConnection implements db reconnection.
func (conn *MongoConn) checkConnection() {
	if err := conn.Ping(); err != nil {
		Impl.Errf("Failed to ping connection - %q", conn.addr)
		Impl.Debugf("Reconnect logic unimplemented")
	}
}

func closeSession(ctx context.Context, session mongo.Session) error {
	session.EndSession(ctx)

	err := session.Client().Disconnect(ctx)
	if err != nil {
		return errs.Wrapf(err, "failed to close session client connection")
	}

	return nil
}

func createConnKey(uri uri.URI, params map[string]string) connKey {
	tlsType := params[tlsConnectParam]
	if tlsType == "" {
		tlsType = disable
	}

	return connKey{
		uri:        uri,
		rawUri:     params[uriParam],
		tlsConnect: tlsType,
		tlsCA:      params[tlsCAParam],
		tlsCert:    params[tlsCertParam],
		tlsKey:     params[tlsKeyParam],
	}
}
