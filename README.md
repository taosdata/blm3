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
         Server "127.0.0.1" "6045"
</Plugin>
```

### statsd

statsd modify the configuration file `path_to_statsd/config.js`

* > `backends` add `"./backends/repeater"`
* > `repeater` add `{ host:'host to blm3', port: 6044}`

example config

```
{
port: 8125
, backends: ["./backends/repeater"]
, repeater: [{ host: '127.0.0.1', port: 6044}]
}
```

## icinga2 restful

gather services & hosts status using icinga2 remote api

* Follow the doc to open icinga2 remote
  api [https://icinga.com/docs/icinga-2/latest/doc/12-icinga2-api/](https://icinga.com/docs/icinga-2/latest/doc/12-icinga2-api/)
* Configure blm3 `icinga2` related parameters

## icinga2 opentsdb writer

collect check result metrics and performance data

* Follow the doc to enable
  opentsdb-writer [https://icinga.com/docs/icinga-2/latest/doc/14-features/#opentsdb-writer](https://icinga.com/docs/icinga-2/latest/doc/14-features/#opentsdb-writer)
* Enable blm3 configuration `opentsdb_telnet.enable`
* Modify the configuration file `/etc/icinga2/features-enabled/opentsdb.conf`

```
object OpenTsdbWriter "opentsdb" {
  host = "host to blm3"
  port = 6046
}
```

## TCollector

tcollector is a client-side process that gathers data from local collectors and pushes the data to OpenTSDB. You run it
on all your hosts, and it does the work of sending each host’s data to the TSD.

* Enable blm3 configuration `opentsdb_telnet.enable`
* Modify the TCollector configuration file, modify the opentsdb host to the host where blm is deployed, and modify the
  port to 6046

## node_exporter

exporter for hardware and OS metrics exposed by *NIX kernels

* Enable blm3 configuration `node_exporter.enable`
* Set the relevant configuration of node_exporter
* Restart blm3

## Configuration

Support command line parameters, environment variables and configuration files
`Command line parameters take precedence over environment variables take precedence over configuration files`
The command line usage is arg=val such as `blm3 -p=30000 --debug=true`

```shell
Usage of blm3:
      --collectd.db string                           collectd db name. Env "BLM_COLLECTD_DB" (default "collectd")
      --collectd.enable                              enable collectd. Env "BLM_COLLECTD_ENABLE" (default true)
      --collectd.password string                     collectd password. Env "BLM_COLLECTD_PASSWORD" (default "taosdata")
      --collectd.port int                            collectd server port. Env "BLM_COLLECTD_PORT" (default 6045)
      --collectd.user string                         collectd user. Env "BLM_COLLECTD_USER" (default "root")
      --collectd.worker int                          collectd write worker. Env "BLM_COLLECTD_WORKER" (default 10)
  -c, --config string                                config path default /etc/taos/blm.toml
      --cors.allowAllOrigins                         cors allow all origins. Env "BLM_CORS_ALLOW_ALL_ORIGINS"
      --cors.allowCredentials                        cors allow credentials. Env "BLM_CORS_ALLOW_Credentials"
      --cors.allowHeaders stringArray                cors allow HEADERS. Env "BLM_ALLOW_HEADERS"
      --cors.allowOrigins stringArray                cors allow origins. Env "BLM_ALLOW_ORIGINS"
      --cors.allowWebSockets                         cors allow WebSockets. Env "BLM_CORS_ALLOW_WebSockets"
      --cors.exposeHeaders stringArray               cors expose headers. Env "BLM_Expose_Headers"
      --debug                                        enable debug mode. Env "BLM_DEBUG"
      --help                                         Print this help message and exit
      --icinga2.caCertFile string                    icinga2 ca cert file path. Env "BLM_ICINGA2_CA_CERT_FILE"
      --icinga2.certFile string                      icinga2 cert file path. Env "BLM_ICINGA2_CERT_FILE"
      --icinga2.db string                            icinga2 db name. Env "BLM_ICINGA2_DB" (default "icinga2")
      --icinga2.enable                               enable icinga2. Env "BLM_ICINGA2_ENABLE"
      --icinga2.gatherDuration duration              icinga2 gather duration. Env "BLM_ICINGA2_GATHER_DURATION" (default 5s)
      --icinga2.host string                          icinga2 server restful host. Env "BLM_ICINGA2_HOST"
      --icinga2.httpPassword string                  icinga2 http password. Env "BLM_ICINGA2_HTTP_PASSWORD"
      --icinga2.httpUsername string                  icinga2 http username. Env "BLM_ICINGA2_HTTP_USERNAME"
      --icinga2.insecureSkipVerify                   icinga2 skip ssl check. Env "BLM_ICINGA2_INSECURE_SKIP_VERIFY" (default true)
      --icinga2.keyFile string                       icinga2 cert key file path. Env "BLM_ICINGA2_KEY_FILE"
      --icinga2.password string                      icinga2 password. Env "BLM_ICINGA2_PASSWORD" (default "taosdata")
      --icinga2.responseTimeout duration             icinga2 response timeout. Env "BLM_ICINGA2_RESPONSE_TIMEOUT" (default 5s)
      --icinga2.user string                          icinga2 user. Env "BLM_ICINGA2_USER" (default "root")
      --influxdb.enable                              enable influxdb. Env "BLM_INFLUXDB_ENABLE" (default true)
      --log.path string                              log path. Env "BLM_LOG_PATH" (default "/var/log/taos")
      --log.rotationCount uint                       log rotation count. Env "BLM_LOG_ROTATION_COUNT" (default 30)
      --log.rotationSize string                      log rotation size(KB MB GB), must be a positive integer. Env "BLM_LOG_ROTATION_SIZE" (default "1GB")
      --log.rotationTime duration                    log rotation time. Env "BLM_LOG_ROTATION_TIME" (default 24h0m0s)
      --logLevel string                              log level (panic fatal error warn warning info debug trace). Env "BLM_LOG_LEVEL" (default "info")
      --node_exporter.caCertFile string              node_exporter ca cert file path. Env "BLM_NODE_EXPORTER_CA_CERT_FILE"
      --node_exporter.certFile string                node_exporter cert file path. Env "BLM_NODE_EXPORTER_CERT_FILE"
      --node_exporter.db string                      node_exporter db name. Env "BLM_NODE_EXPORTER_DB" (default "node_exporter")
      --node_exporter.enable                         enable node_exporter. Env "BLM_NODE_EXPORTER_ENABLE"
      --node_exporter.gatherDuration duration        node_exporter gather duration. Env "BLM_NODE_EXPORTER_GATHER_DURATION" (default 5s)
      --node_exporter.httpBearerTokenString string   node_exporter http bearer token. Env "BLM_NODE_EXPORTER_HTTP_BEARER_TOKEN_STRING"
      --node_exporter.httpPassword string            node_exporter http password. Env "BLM_NODE_EXPORTER_HTTP_PASSWORD"
      --node_exporter.httpUsername string            node_exporter http username. Env "BLM_NODE_EXPORTER_HTTP_USERNAME"
      --node_exporter.insecureSkipVerify             node_exporter skip ssl check. Env "BLM_NODE_EXPORTER_INSECURE_SKIP_VERIFY" (default true)
      --node_exporter.keyFile string                 node_exporter cert key file path. Env "BLM_NODE_EXPORTER_KEY_FILE"
      --node_exporter.password string                node_exporter password. Env "BLM_NODE_EXPORTER_PASSWORD" (default "taosdata")
      --node_exporter.responseTimeout duration       node_exporter response timeout. Env "BLM_NODE_EXPORTER_RESPONSE_TIMEOUT" (default 5s)
      --node_exporter.urls strings                   node_exporter urls. Env "BLM_NODE_EXPORTER_URLS"
      --node_exporter.user string                    node_exporter user. Env "BLM_NODE_EXPORTER_USER" (default "root")
      --opentsdb.enable                              enable opentsdb. Env "BLM_OPENTSDB_ENABLE" (default true)
      --opentsdb_telnet.db string                    opentsdb_telnet db name. Env "BLM_OPENTSDB_TELNET_DB" (default "opentsdb_telnet")
      --opentsdb_telnet.enable                       enable opentsdb telnet,warning: without auth info(default false). Env "BLM_OPENTSDB_TELNET_ENABLE"
      --opentsdb_telnet.maxTCPConnections int        max tcp connections. Env "BLM_OPENTSDB_TELNET_MAX_TCP_CONNECTIONS" (default 250)
      --opentsdb_telnet.password string              opentsdb_telnet password. Env "BLM_OPENTSDB_TELNET_PASSWORD" (default "taosdata")
      --opentsdb_telnet.port int                     opentsdb telnet tcp port. Env "BLM_OPENTSDB_TELNET_PORT" (default 6046)
      --opentsdb_telnet.tcpKeepAlive                 enable tcp keep alive. Env "BLM_OPENTSDB_TELNET_TCP_KEEP_ALIVE"
      --opentsdb_telnet.user string                  opentsdb_telnet user. Env "BLM_OPENTSDB_TELNET_USER" (default "root")
      --opentsdb_telnet.worker int                   opentsdb_telnet write worker. Env "BLM_OPENTSDB_TELNET_WORKER" (default 1000)
      --pool.maxConnect int                          max connections to taosd. Env "BLM_POOL_MAX_CONNECT" (default 4000)
      --pool.maxIdle int                             max idle connections to taosd. Env "BLM_POOL_MAX_IDLE" (default 4000)
  -P, --port int                                     http port. Env "BLM_PORT" (default 6041)
      --ssl.certFile string                          ssl cert file path. Env "BLM_SSL_CERT_FILE"
      --ssl.enable                                   enable ssl. Env "BLM_SSL_ENABLE"
      --ssl.keyFile string                           ssl key file path. Env "BLM_SSL_KEY_FILE"
      --statsd.allowPendingMessages int              statsd allow pending messages. Env "BLM_STATSD_ALLOW_PENDING_MESSAGES" (default 50000)
      --statsd.db string                             statsd db name. Env "BLM_STATSD_DB" (default "statsd")
      --statsd.deleteCounters                        statsd delete counter cache after gather. Env "BLM_STATSD_DELETE_COUNTERS" (default true)
      --statsd.deleteGauges                          statsd delete gauge cache after gather. Env "BLM_STATSD_DELETE_GAUGES" (default true)
      --statsd.deleteSets                            statsd delete set cache after gather. Env "BLM_STATSD_DELETE_SETS" (default true)
      --statsd.deleteTimings                         statsd delete timing cache after gather. Env "BLM_STATSD_DELETE_TIMINGS" (default true)
      --statsd.enable                                enable statsd. Env "BLM_STATSD_ENABLE" (default true)
      --statsd.gatherInterval duration               statsd gather interval. Env "BLM_STATSD_GATHER_INTERVAL" (default 5s)
      --statsd.maxTCPConnections int                 statsd max tcp connections. Env "BLM_STATSD_MAX_TCP_CONNECTIONS" (default 250)
      --statsd.password string                       statsd password. Env "BLM_STATSD_PASSWORD" (default "taosdata")
      --statsd.port int                              statsd server port. Env "BLM_STATSD_PORT" (default 6044)
      --statsd.protocol string                       statsd protocol [tcp or udp]. Env "BLM_STATSD_PROTOCOL" (default "udp")
      --statsd.tcpKeepAlive                          enable tcp keep alive. Env "BLM_STATSD_TCP_KEEP_ALIVE"
      --statsd.user string                           statsd user. Env "BLM_STATSD_USER" (default "root")
      --statsd.worker int                            statsd write worker. Env "BLM_STATSD_WORKER" (default 10)
      --taosConfigDir string                         load taos client config path. Env "BLM_TAOS_CONFIG_FILE"
      --version                                      Print the version and exit
```

For the default configuration file, see [example/config/blm.toml](example/config/blm.toml)
