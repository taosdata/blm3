# BLM3

## Function

* Compatible with restful interface
* Compatible with influxdb v1 write interface
* Compatible with opentsdb json and telnet format writing
* Seamless connection collectd
* Seamless connection with statsd

## Interface

### restful

```
/rest/sql
/rest/sqlt
/rest/sqlutc
```

### influxdb

```
/influxdb/v1/write
```

Support query parameters
> `db` Specify the necessary parameters for the database
> `precision` time precision non-essential parameter
> `u` user non-essential parameters
> `p` password Optional parameter

### opentsdb

```
/opentsdb/v1/put/json/:db
/opentsdb/v1/put/telnet/:db
```

### collectd

Modify the collected configuration `/etc/collectd/collectd.conf`

```
LoadPlugin network
<Plugin network>
         Server "127.0.0.1" "6043"
</Plugin>
```

### statsd

statsd modify the configuration file `path_to_statsd/config.js`

* > `backends` add `"./backends/repeater"`
* > `repeater` add `{ host:'host to blm3', port: 6044}`

example config

```
{
port: 6044 
, backends: ["./backends/repeater"]
, repeater: [{ host: '127.0.0.1', port: 6044}]
}
```



## Configuration

Support command line parameters, environment variables and configuration files
`Command line parameters take precedence over environment variables take precedence over configuration files`
The command line usage is arg=val such as `blm3 -p=30000 --debug=true`

```shell
Usage of blm3:
      --collectd.db string                collectd db name. Env "BLM_COLLECTD_DB" (default "collectd")
      --collectd.enable                   enable collectd. Env "BLM_COLLECTD_ENABLE" (default true)
      --collectd.password string          collectd password. Env "BLM_COLLECTD_PASSWORD" (default "taosdata")
      --collectd.port int                 collectd server port. Env "BLM_COLLECTD_PORT" (default 6043)
      --collectd.user string              collectd user. Env "BLM_COLLECTD_USER" (default "root")
      --collectd.worker int               collectd write worker. Env "BLM_COLLECTD_WORKER" (default 10)
  -c, --config string                     config path default /etc/taos/blm.toml
      --cors.allowAllOrigins              cors allow all origins. Env "BLM_CORS_ALLOW_ALL_ORIGINS"
      --cors.allowCredentials             cors allow credentials. Env "BLM_CORS_ALLOW_Credentials"
      --cors.allowHeaders stringArray     cors allow HEADERS. Env "BLM_ALLOW_HEADERS"
      --cors.allowOrigins stringArray     cors allow origins. Env "BLM_ALLOW_ORIGINS"
      --cors.allowWebSockets              cors allow WebSockets. Env "BLM_CORS_ALLOW_WebSockets"
      --cors.exposeHeaders stringArray    cors expose headers. Env "BLM_Expose_Headers"
      --debug                             enable debug mode. Env "BLM_DEBUG"
      --influxdb.enable                   enable influxdb. Env "BLM_INFLUXDB_ENABLE" (default true)
      --log.path string                   log path. Env "BLM_LOG_PATH" (default "/var/log/taos")
      --log.rotationCount uint            log rotation count. Env "BLM_LOG_ROTATION_COUNT" (default 30)
      --log.rotationSize string           log rotation size(KB MB GB), must be a positive integer. Env "BLM_LOG_ROTATION_SIZE" (default "1GB")
      --log.rotationTime duration         log rotation time. Env "BLM_LOG_ROTATION_TIME" (default 24h0m0s)
      --logLevel string                   log level (panic fatal error warn warning info debug trace). Env "BLM_LOG_LEVEL" (default "info")
      --opentsdb.enable                   enable opentsdb. Env "BLM_OPENTSDB_ENABLE" (default true)
      --pool.idleTimeout duration         Set idle connection timeout. Env "BLM_POOL_IDLE_TIMEOUT" (default 1m0s)
      --pool.maxConnect int               max connections to taosd. Env "BLM_POOL_MAX_CONNECT" (default 100)
      --pool.maxIdle int                  max idle connections to taosd. Env "BLM_POOL_MAX_IDLE" (default 5)
  -P, --port int                          http port. Env "BLM_PORT" (default 6041)
      --ssl.certFile string               ssl cert file path. Env "BLM_SSL_CERT_FILE"
      --ssl.enable                        enable ssl. Env "BLM_SSL_ENABLE"
      --ssl.keyFile string                ssl key file path. Env "BLM_SSL_KEY_FILE"
      --statsd.allowPendingMessages int   statsd allow pending messages. Env "BLM_STATSD_ALLOW_PENDING_MESSAGES" (default 50000)
      --statsd.db string                  statsd db name. Env "BLM_STATSD_DB" (default "statsd")
      --statsd.deleteCounters             statsd delete counter cache after gather. Env "BLM_STATSD_DELETE_COUNTERS" (default true)
      --statsd.deleteGauges               statsd delete gauge cache after gather. Env "BLM_STATSD_DELETE_GAUGES" (default true)
      --statsd.deleteSets                 statsd delete set cache after gather. Env "BLM_STATSD_DELETE_SETS" (default true)
      --statsd.deleteTimings              statsd delete timing cache after gather. Env "BLM_STATSD_DELETE_TIMINGS" (default true)
      --statsd.enable                     enable statsd. Env "BLM_STATSD_ENABLE" (default true)
      --statsd.gatherInterval duration    statsd gather interval. Env "BLM_STATSD_GATHER_INTERVAL" (default 5s)
      --statsd.maxTCPConnections int      statsd max tcp connections. Env "BLM_STATSD_MAX_TCP_CONNECTIONS" (default 250)
      --statsd.password string            statsd password. Env "BLM_STATSD_PASSWORD" (default "taosdata")
      --statsd.port int                   statsd server port. Env "BLM_STATSD_PORT" (default 6044)
      --statsd.protocol string            statsd protocol [tcp or udp]. Env "BLM_STATSD_PROTOCOL" (default "udp")
      --statsd.tcpKeepAlive               enable tcp keep alive. Env "BLM_COLLECTD_TCP_KEEP_ALIVE"
      --statsd.user string                statsd user. Env "BLM_STATSD_USER" (default "root")
      --statsd.worker int                 statsd write worker. Env "BLM_STATSD_WORKER" (default 10)
```

For the default configuration file, see [example/config/blm.toml](example/config/blm.toml)
