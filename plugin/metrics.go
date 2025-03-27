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
	"golang.zabbix.com/plugin/mongodb/plugin/handlers"
	"golang.zabbix.com/sdk/metric"
	"golang.zabbix.com/sdk/plugin"
	"golang.zabbix.com/sdk/uri"
)

const (
	keyConfigDiscovery      = "mongodb.cfg.discovery"
	keyCollectionStats      = "mongodb.collection.stats"
	keyCollectionsDiscovery = "mongodb.collections.discovery"
	keyCollectionsUsage     = "mongodb.collections.usage"
	keyConnPoolStats        = "mongodb.connpool.stats"
	keyDatabaseStats        = "mongodb.db.stats"
	keyDatabasesDiscovery   = "mongodb.db.discovery"
	keyJumboChunks          = "mongodb.jumbo_chunks.count"
	keyOplogStats           = "mongodb.oplog.stats"
	keyPing                 = "mongodb.ping"
	keyReplSetConfig        = "mongodb.rs.config"
	keyReplSetStatus        = "mongodb.rs.status"
	keyServerStatus         = "mongodb.server.status"
	keyShardsDiscovery      = "mongodb.sh.discovery"
	keyVersion              = "mongodb.version"

	uriParam        = "URI"
	tlsConnectParam = "TLSConnect"
	tlsCAParam      = "TLSCAFile"
	tlsCertParam    = "TLSCertFile"
	tlsKeyParam     = "TLSKeyFile"
)

var metricHandlers = map[string]handlerFunc{
	keyCollectionStats:      handlers.CollectionStatsHandler,
	keyCollectionsDiscovery: handlers.CollectionsDiscoveryHandler,
	keyCollectionsUsage:     handlers.CollectionsUsageHandler,
	keyConfigDiscovery:      handlers.ConfigDiscoveryHandler,
	keyConnPoolStats:        handlers.ConnPoolStatsHandler,
	keyDatabaseStats:        handlers.DatabaseStatsHandler,
	keyDatabasesDiscovery:   handlers.DatabasesDiscoveryHandler,
	keyJumboChunks:          handlers.JumboChunksHandler,
	keyOplogStats:           handlers.OplogStatsHandler,
	keyPing:                 handlers.PingHandler,
	keyReplSetConfig:        handlers.ReplSetConfigHandler,
	keyReplSetStatus:        handlers.ReplSetStatusHandler,
	keyServerStatus:         handlers.ServerStatusHandler,
	keyShardsDiscovery:      handlers.ShardsDiscoveryHandler,
	keyVersion:              handlers.VersionHandler,
}

var (
	paramURI = metric.NewConnParam(uriParam, "URI to connect or session name.").
			WithDefault(handlers.UriDefaults.Scheme + "://localhost:" + handlers.UriDefaults.Port).WithSession().
			WithValidator(uri.URIValidator{Defaults: handlers.UriDefaults, AllowedSchemes: []string{"tcp"}})
	paramUser        = metric.NewConnParam("User", "MongoDB user.")
	paramPassword    = metric.NewConnParam("Password", "User's password.")
	paramDatabase    = metric.NewParam("Database", "Database name.").WithDefault("admin")
	paramCollection  = metric.NewParam("Collection", "Collection name.").SetRequired()
	paramTLSConnect  = metric.NewSessionOnlyParam(tlsConnectParam, "DB connection encryption type.").WithDefault("")
	paramTLSCaFile   = metric.NewSessionOnlyParam(tlsCAParam, "TLS ca file path.").WithDefault("")
	paramTLSCertFile = metric.NewSessionOnlyParam(tlsCertParam, "TLS cert file path.").WithDefault("")
	paramTLSKeyFile  = metric.NewSessionOnlyParam(tlsKeyParam, "TLS key file path.").WithDefault("")
)

var metrics = metric.MetricSet{
	keyCollectionStats: metric.New(
		"Returns a variety of storage statistics for a given collection.",
		[]*metric.Param{
			paramURI, paramUser, paramPassword, paramDatabase, paramCollection,
			paramTLSConnect, paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile,
		},
		false,
	),

	keyCollectionsDiscovery: metric.New(
		"Returns a list of discovered collections.",
		[]*metric.Param{
			paramURI, paramUser, paramPassword,
			paramTLSConnect, paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile,
		},
		false,
	),

	keyCollectionsUsage: metric.New(
		"Returns usage statistics for collections.",
		[]*metric.Param{
			paramURI, paramUser, paramPassword,
			paramTLSConnect, paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile,
		},
		false,
	),

	keyConfigDiscovery: metric.New(
		"Returns a list of discovered config servers.",
		[]*metric.Param{
			paramURI, paramUser, paramPassword, paramTLSConnect, paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile,
		},
		false,
	),

	keyConnPoolStats: metric.New(
		"Returns information regarding the open outgoing connections from the "+
			"current database instance to other members of the sharded cluster or replica set.",
		[]*metric.Param{
			paramURI, paramUser, paramPassword,
			paramTLSConnect, paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile,
		},
		false,
	),

	keyDatabaseStats: metric.New(
		"Returns statistics reflecting a given database system’s state.",
		[]*metric.Param{
			paramURI, paramUser, paramPassword, paramDatabase,
			paramTLSConnect, paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile,
		},
		false,
	),

	keyDatabasesDiscovery: metric.New(
		"Returns a list of discovered databases.",
		[]*metric.Param{
			paramURI, paramUser, paramPassword,
			paramTLSConnect, paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile,
		},
		false,
	),

	keyJumboChunks: metric.New(
		"Returns count of jumbo chunks.",
		[]*metric.Param{
			paramURI, paramUser, paramPassword,
			paramTLSConnect, paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile,
		},
		false,
	),

	keyOplogStats: metric.New(
		"Returns a status of the replica set, using data polled from the oplog.",
		[]*metric.Param{
			paramURI, paramUser, paramPassword,
			paramTLSConnect, paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile,
		},
		false,
	),

	keyPing: metric.New(
		"Test if connection is alive or not.",
		[]*metric.Param{
			paramURI, paramUser, paramPassword,
			paramTLSConnect, paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile,
		},
		false,
	),

	keyReplSetConfig: metric.New(
		"Returns a current configuration of the replica set.",
		[]*metric.Param{
			paramURI, paramUser, paramPassword,
			paramTLSConnect, paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile,
		},
		false,
	),

	keyReplSetStatus: metric.New(
		"Returns a replica set status from the point of view of the member "+
			"where the method is run.",
		[]*metric.Param{
			paramURI, paramUser, paramPassword,
			paramTLSConnect, paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile,
		},
		false,
	),

	keyServerStatus: metric.New(
		"Returns a database’s state.",
		[]*metric.Param{
			paramURI, paramUser, paramPassword,
			paramTLSConnect, paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile,
		},
		false,
	),

	keyShardsDiscovery: metric.New(
		"Returns a list of discovered shards present in the cluster.",
		[]*metric.Param{
			paramURI, paramUser, paramPassword,
			paramTLSConnect, paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile,
		},
		false,
	),

	keyVersion: metric.New(
		"Returns database version.",
		[]*metric.Param{
			paramURI, paramUser, paramPassword,
			paramTLSConnect, paramTLSCaFile, paramTLSCertFile, paramTLSKeyFile,
		},
		false,
	),
}

// handlerFunc defines an interface must be implemented by handlers.
type handlerFunc func(s handlers.Session, params map[string]string) (res interface{}, err error)

func init() {
	err := plugin.RegisterMetrics(&Impl, Name, metrics.List()...)
	if err != nil {
		panic(err)
	}
}

// getHandlerFunc returns a handlerFunc related to a given key.
func getHandlerFunc(key string) handlerFunc {
	return metricHandlers[key]
}
