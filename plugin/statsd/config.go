package statsd

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/taosdata/driver-go/v2/common"
	"time"
)

type Config struct {
	Enable                 bool
	Port                   int
	DB                     string
	User                   string
	Password               string
	Worker                 int
	GatherInterval         time.Duration
	Protocol               string
	MaxTCPConnections      int
	TCPKeepAlive           bool
	AllowedPendingMessages int
	DeleteCounters         bool
	DeleteGauges           bool
	DeleteSets             bool
	DeleteTimings          bool
}

func (c *Config) setValue() {
	c.Enable = viper.GetBool("statsd.enable")
	c.Port = viper.GetInt("statsd.port")
	c.DB = viper.GetString("statsd.db")
	c.User = viper.GetString("statsd.user")
	c.Password = viper.GetString("statsd.password")
	c.Worker = viper.GetInt("statsd.worker")
	c.GatherInterval = viper.GetDuration("statsd.gatherInterval")
	c.Protocol = viper.GetString("statsd.protocol")
	c.MaxTCPConnections = viper.GetInt("statsd.maxTCPConnections")
	c.TCPKeepAlive = viper.GetBool("statsd.tcpKeepAlive")
	c.AllowedPendingMessages = viper.GetInt("statsd.allowPendingMessages")
	c.DeleteCounters = viper.GetBool("statsd.deleteCounters")
	c.DeleteGauges = viper.GetBool("statsd.deleteGauges")
	c.DeleteSets = viper.GetBool("statsd.deleteSets")
	c.DeleteTimings = viper.GetBool("statsd.deleteTimings")
}

func init() {
	_ = viper.BindEnv("statsd.enable", "BLM_STATSD_ENABLE")
	pflag.Bool("statsd.enable", true, `enable statsd. Env "BLM_STATSD_ENABLE"`)
	viper.SetDefault("statsd.enable", true)

	_ = viper.BindEnv("statsd.port", "BLM_STATSD_PORT")
	pflag.Int("statsd.port", 8125, `statsd server port. Env "BLM_STATSD_PORT"`)
	viper.SetDefault("statsd.port", 8125)

	_ = viper.BindEnv("statsd.db", "BLM_STATSD_DB")
	pflag.String("statsd.db", "statsd", `statsd db name. Env "BLM_STATSD_DB"`)
	viper.SetDefault("statsd.db", "statsd")

	_ = viper.BindEnv("statsd.user", "BLM_STATSD_USER")
	pflag.String("statsd.user", common.DefaultUser, `statsd user. Env "BLM_STATSD_USER"`)
	viper.SetDefault("statsd.user", common.DefaultUser)

	_ = viper.BindEnv("statsd.password", "BLM_STATSD_PASSWORD")
	pflag.String("statsd.password", common.DefaultPassword, `statsd password. Env "BLM_STATSD_PASSWORD"`)
	viper.SetDefault("statsd.password", common.DefaultPassword)

	_ = viper.BindEnv("statsd.worker", "BLM_STATSD_WORKER")
	pflag.Int("statsd.worker", 10, `statsd write worker. Env "BLM_STATSD_WORKER"`)
	viper.SetDefault("statsd.worker", 10)

	_ = viper.BindEnv("statsd.gatherInterval", "BLM_STATSD_GATHER_INTERVAL")
	pflag.Duration("statsd.gatherInterval", time.Second*30, `statsd gather interval. Env "BLM_STATSD_GATHER_INTERVAL"`)
	viper.SetDefault("statsd.gatherInterval", "30s")

	_ = viper.BindEnv("statsd.protocol", "BLM_STATSD_PROTOCOL")
	pflag.String("statsd.protocol", "udp", `statsd protocol [tcp or udp]. Env "BLM_STATSD_PROTOCOL"`)
	viper.SetDefault("statsd.protocol", "udp")

	_ = viper.BindEnv("statsd.maxTCPConnections", "BLM_STATSD_MAX_TCP_CONNECTIONS")
	pflag.Int("statsd.maxTCPConnections", 250, `statsd max tcp connections. Env "BLM_STATSD_MAX_TCP_CONNECTIONS"`)
	viper.SetDefault("statsd.maxTCPConnections", 250)

	_ = viper.BindEnv("statsd.tcpKeepAlive", "BLM_COLLECTD_TCP_KEEP_ALIVE")
	pflag.Bool("statsd.tcpKeepAlive", false, `enable tcp keep alive. Env "BLM_COLLECTD_TCP_KEEP_ALIVE"`)
	viper.SetDefault("statsd.tcpKeepAlive", false)

	_ = viper.BindEnv("statsd.allowPendingMessages", "BLM_STATSD_ALLOW_PENDING_MESSAGES")
	pflag.Int("statsd.allowPendingMessages", 10000, `statsd allow pending messages. Env "BLM_STATSD_ALLOW_PENDING_MESSAGES"`)
	viper.SetDefault("statsd.allowPendingMessages", 10000)

	_ = viper.BindEnv("statsd.deleteCounters", "BLM_STATSD_DELETE_COUNTERS")
	pflag.Bool("statsd.deleteCounters", true, `statsd delete counter cache after gather. Env "BLM_STATSD_DELETE_COUNTERS"`)
	viper.SetDefault("statsd.deleteCounters", true)

	_ = viper.BindEnv("statsd.deleteGauges", "BLM_STATSD_DELETE_GAUGES")
	pflag.Bool("statsd.deleteGauges", true, `statsd delete gauge cache after gather. Env "BLM_STATSD_DELETE_GAUGES"`)
	viper.SetDefault("statsd.deleteGauges", true)

	_ = viper.BindEnv("statsd.deleteSets", "BLM_STATSD_DELETE_SETS")
	pflag.Bool("statsd.deleteSets", true, `statsd delete set cache after gather. Env "BLM_STATSD_DELETE_SETS"`)
	viper.SetDefault("statsd.deleteSets", true)

	_ = viper.BindEnv("statsd.deleteTimings", "BLM_STATSD_DELETE_TIMINGS")
	pflag.Bool("statsd.deleteTimings", true, `statsd delete timing cache after gather. Env "BLM_STATSD_DELETE_TIMINGS"`)
	viper.SetDefault("statsd.deleteTimings", true)
}