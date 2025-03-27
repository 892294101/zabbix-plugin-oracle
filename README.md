# MongoDB plugin
This plugin provides a native Zabbix solution to monitor MongoDB servers and clusters (document-based, distributed database). 
It can monitor several MongoDB instances simultaneously, remotes or locals to the Zabbix Agent. 
The plugin keeps connections in the opened state to reduce network 
congestion, latency, CPU and memory usage. It is best for use in conjunction with the official 
[MongoDB template.](https://git.zabbix.com/projects/ZBX/repos/zabbix/browse/templates/app/mongodb)
You can extend it or create your own template to cater specific needs.

## Requirements
* Zabbix Agent 2 version 6.0.0 or newer
* Go >= 1.21 (required only to build from the source)

## Supported versions
* MongoDB, versions from 2.6 till 5.3

## Plugin setup
*Plugins.MongoDB.System.Path* variable needs to be set in Zabbix agent 2 configuration file with the path to the MongoDB
plugin executable. By default the variable is set in **plugin** configuration file *mongodb.conf* and then included in
the **agent** configuration file *zabbix_agent2.conf*.

For example: 
You should add the following option to the **plugin** configuration file:

    Plugins.MongoDB.System.Path=/path/to/executable/mongodb

Then, the configuration file needs to be included in the main Zabbix agent 2 configuration file via the
*Include* command.

For example: 
You should add the following option to the **plugin** configuration file:

    Include=/path/to/config/mongodb.conf

## Options
MongoDB plugin can be executed on its own with these parameters:
* *-h*, *--help* displays help message;
* *-V*, *--version* displays the plugin version and license information.

## Installation
Depending on your configuration you need to create a local read-only user in the admin database:  
- *STANDALONE*: for each single MongoDB node;
- *REPLICASET*: create the user on the primary node of the replica set;
- *SHARDING*: for each shard in your cluster (just create the user on the primary node of the replica set). 
Also, create the same user on a mongos router. It will automatically spread to configuration servers.

```javascript
use admin

db.auth("admin", "<ADMIN_PASSWORD>")

db.createUser({
  "user": "zabbix",
  "pwd": "<PASSWORD>",
  "roles": [
    { role: "readAnyDatabase", db: "admin" },
    { role: "clusterMonitor", db: "admin" },
  ]
})
```

## Configuration
To configure plugins, use Zabbix agent configuration file.

**Plugins.MongoDB.KeepAlive** — sets the time for waiting before unused connections will be closed.  
*Default value:* 300 sec.  
*Limits:* 60-900

**Plugins.MongoDB.Timeout** — the amount of time to wait for a server to respond when connecting for the first time and on follow up 
operations in the session.  
*Default value:* equals the global Timeout configuration parameter.  
*Limits:* 1-30

**Plugins.MongoDB.Sessions.<session_name>.TLSConnect** — encryption type for MongoDB connection. 
"*" should be replaced with a session name. 
*Default value:* empty
Accepted values: required, verify_ca, verify_full

**Plugins.MongoDB.Sessions.<session_name>.TLSCAFile** — full pathname of a file containing the 
top-level CA(s) certificates for MongoDB. 
*Default value:* empty

**Plugins.MongoDB.Sessions.<session_name>.TLSCertFile** — full pathname of a file containing the MongoDB certificate or certificate chain. 
*Default value:* empty

**Plugins.MongoDB.Sessions.*.TLSKeyFile** — full pathname of a file containing the MongoDB private key. 
*Default value:* empty

### Configuring connection
A connection can be configured using either keys' parameters or named sessions.     

*Notes*:  
* You can leave any connection parameter empty, a default hard-coded value will be used in such case: 
  localhost:27017 without authentication.
* Embedded URI credentials (userinfo) are forbidden and will be ignored. 

Thus, you cannot pass the credentials by this:   
  
      mongodb.ping[tcp://user:password@127.0.0.1] — WRONG  
  
  The correct way is:
    
      mongodb.ping[tcp://127.0.0.1,user,password] - CORRECT
      
* Currently, only TCP connections are supported.
  
These are examples of valid URIs:
    - tcp://127.0.0.1:27017
    - tcp://localhost
    - localhost
      
#### Using keys' parameters
The common parameters for all keys are: [ConnString][,User][,Password].  
Where *ConnString* can be either a URI or session name.   
*ConnString* will be treated as a URI if no session with the given name are found.  
If you use *ConnString* as a session name, just skip the rest of the connection parameters.  
 
#### Using named sessions
Named sessions allow you to define specific parameters for each MongoDB instance. 
Currently, these are the supported parameters: Uri, User, Password, TLSConnect, TLSCAFile,
TLSCertFile and TLSKeyFile.
It's is a more secure way to store credentials compared to item keys or macros.  

For example, if you have two MongoDB instances: "Prod" and "Test", 
you should add the following options to the plugin configuration file:   

    Plugins.MongoDB.Sessions.Prod.Uri=tcp://192.168.1.1:27017
    Plugins.MongoDB.Sessions.Prod.User=<UserForProd>
    Plugins.MongoDB.Sessions.Prod.Password=<PasswordForProd>
    Plugins.MongoDB.Sessions.Prod.TLSConnect=verify_full
    Plugins.MongoDB.Sessions.Prod.TLSCAFile=/path/to/ca_file
    Plugins.MongoDB.Sessions.Prod.TLSCertFile=/path/to/cert_file
    Plugins.MongoDB.Sessions.Prod.TLSKeyFile=/path/to/key_file

    Plugins.MongoDB.Sessions.Test.Uri=tcp://192.168.0.1:27017
    Plugins.MongoDB.Sessions.Test.User=<UserForTest>
    Plugins.MongoDB.Sessions.Test.Password=<PasswordForTest>
    Plugins.MongoDB.Sessions.Test.TLSConnect=verify_ca
    Plugins.MongoDB.Sessions.Test.TLSCAFile=/path/to/ca_file
    Plugins.MongoDB.Sessions.Test.TLSCertFile=/path/to/cert_file
    Plugins.MongoDB.Sessions.Test.TLSKeyFile=/path/to/key_file
        
Then, you will be able to use these names as the first parameter (ConnString) in keys instead of URIs.

For example:

    mongodb.ping[Prod]
    mongodb.ping[Test]

*Note*: sessions names are case-sensitive.

## Supported keys
**mongodb.collection.stats[\<commonParams\>[,database],collection]** — returns a variety of storage statistics for a 
given collection.  
*Parameters:*  
database — database name (default: admin).  
collection (required) — collection name.

**mongodb.cfg.discovery[\<commonParams\>]** — returns a list of discovered configuration servers.  

**mongodb.collections.discovery[\<commonParams\>]** — returns a list of discovered collections.  

**mongodb.collections.usage[\<commonParams\>]** — returns usage statistics for collections.  

**mongodb.connpool.stats[\<commonParams\>]** — returns the information regarding the open outgoing connections from the
current database instance to other members of the sharded cluster or replica set.    

**mongodb.db.stats[\<commonParams\>[,database]]** — returns statistics reflecting a given database system’s state.  
*Parameters:*  
database — database name (default: admin).    

**mongodb.db.discovery[\<commonParams\>]** — returns a list of discovered databases.    

**mongodb.jumbo_chunks.count[\<commonParams\>]** — returns a count of jumbo chunks.    

**mongodb.oplog.stats[\<commonParams\>]** — returns the status of the replica set, using data polled from the oplog.    

**mongodb.ping[\<commonParams\>]** — tests if a connection is alive or not.  
*Returns:*
- "1" if a connection is alive.
- "0" if a connection is broken (if there is any error presented including AUTH and configuration issues).

**mongodb.rs.config[\<commonParams\>]** — returns the current configuration of the replica set.    

**mongodb.rs.status[\<commonParams\>]** — returns the status of the replica set - as seen by the member
where the method is run.  
 
**mongodb.server.status[\<commonParams\>]** — returns the state of the database.    

**mongodb.sh.discovery[\<commonParams\>]** — returns a list of discovered shards present in the cluster.    

**mongodb.version[\<commonParams\>]** — returns database server version.

## Troubleshooting
The plugin uses logs of Zabbix agent. You can increase debugging level of Zabbix agent if you need more details about the current situation.
Set the *DebugLevel* configuration option to "5" (extended debugging) in order to turn on verbose log messages.
