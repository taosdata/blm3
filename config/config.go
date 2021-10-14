package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	Cors     CorsConfig
	Debug    bool
	Port     int
	LogLevel string
	SSl      SSl
	Log      Log
	Pool     Pool
}

type SSl struct {
	Enable   bool
	CertFile string
	KeyFile  string
}

func initSSL() {
	viper.SetDefault("ssl.enable", false)
	_ = viper.BindEnv("ssl.enable", "BLM_SSL_ENABLE")
	pflag.Bool("ssl.enable", false, `enable ssl. Env "BLM_SSL_ENABLE"`)

	viper.SetDefault("ssl.certFile", "")
	_ = viper.BindEnv("ssl.certFile", "BLM_SSL_CERT_FILE")
	pflag.String("ssl.certFile", "", `ssl cert file path. Env "BLM_SSL_CERT_FILE"`)

	viper.SetDefault("ssl.keyFile", "")
	_ = viper.BindEnv("ssl.keyFile", "BLM_SSL_KEY_FILE")
	pflag.String("ssl.keyFile", "", `ssl key file path. Env "BLM_SSL_KEY_FILE"`)
}

func (s *SSl) setValue() {
	s.Enable = viper.GetBool("ssl.enable")
	s.CertFile = viper.GetString("ssl.certFile")
	s.KeyFile = viper.GetString("ssl.keyFile")
}

type Pool struct {
	MaxConnect  int
	MaxIdle     int
	IdleTimeout time.Duration
}

func initPool() {
	viper.SetDefault("pool.maxConnect", 100)
	_ = viper.BindEnv("pool.maxConnect", "BLM_POOL_MAX_CONNECT")
	pflag.Int("pool.maxConnect", 100, `max connections to taosd. Env "BLM_POOL_MAX_CONNECT"`)

	viper.SetDefault("pool.maxIdle", 5)
	_ = viper.BindEnv("pool.maxIdle", "BLM_POOL_MAX_IDLE")
	pflag.Int("pool.maxIdle", 5, `max idle connections to taosd. Env "BLM_POOL_MAX_IDLE"`)

	viper.SetDefault("pool.idleTimeout", time.Minute)
	_ = viper.BindEnv("pool.idleTimeout", "BLM_POOL_IDLE_TIMEOUT")
	pflag.Duration("pool.idleTimeout", time.Minute, `Set idle connection timeout. Env "BLM_POOL_IDLE_TIMEOUT"`)

}

func (p *Pool) setValue() {
	p.MaxConnect = viper.GetInt("pool.maxConnect")
	p.MaxIdle = viper.GetInt("pool.maxIdle")
	p.IdleTimeout = viper.GetDuration("pool.idleTimeout")
}

type Log struct {
	Path          string
	RotationCount uint
	RotationTime  time.Duration
	RotationSize  uint
}

func initLog() {
	viper.SetDefault("log.path", "/var/log/taos")
	_ = viper.BindEnv("log.path", "BLM_LOG_PATH")
	pflag.String("log.path", "/var/log/taos", `log path. Env "BLM_LOG_PATH"`)

	viper.SetDefault("log.rotationCount", 30)
	_ = viper.BindEnv("log.rotationCount", "BLM_LOG_ROTATION_COUNT")
	pflag.Uint("log.rotationCount", 30, `log rotation count. Env "BLM_LOG_ROTATION_COUNT"`)

	viper.SetDefault("log.rotationTime", time.Hour*24)
	_ = viper.BindEnv("log.rotationTime", "BLM_LOG_ROTATION_TIME")
	pflag.Duration("log.rotationTime", time.Hour*24, `log rotation time. Env "BLM_LOG_ROTATION_TIME"`)

	viper.SetDefault("log.rotationSize", "1GB")
	_ = viper.BindEnv("log.rotationSize", "BLM_LOG_ROTATION_SIZE")
	pflag.String("log.rotationSize", "1GB", `log rotation size(KB MB GB), must be a positive integer. Env "BLM_LOG_ROTATION_SIZE"`)
}

func (l *Log) setValue() {
	l.Path = viper.GetString("log.path")
	l.RotationCount = viper.GetUint("log.rotationCount")
	l.RotationTime = viper.GetDuration("log.rotationTime")
	l.RotationSize = viper.GetSizeInBytes("log.rotationSize")
}

var (
	Conf *Config
)

func Init() {
	viper.SetConfigType("toml")
	viper.SetConfigName("blm")
	viper.AddConfigPath("/etc/taos")
	cp := pflag.StringP("config", "c", "", "config path default /etc/taos/blm.toml")
	pflag.Parse()
	if *cp != "" {
		viper.SetConfigFile(*cp)
	}
	viper.SetEnvPrefix("blm")
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		panic(err)
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("config file not found")
		} else {
			panic(err)
		}
	}
	Conf = &Config{
		Debug:    viper.GetBool("debug"),
		Port:     viper.GetInt("port"),
		LogLevel: viper.GetString("logLevel"),
	}
	Conf.Log.setValue()
	Conf.Cors.setValue()
	Conf.SSl.setValue()
	Conf.Pool.setValue()
}

//arg > file > env
func init() {
	viper.SetDefault("debug", false)
	_ = viper.BindEnv("debug", "BLM_DEBUG")
	pflag.Bool("debug", false, `enable debug mode. Env "BLM_DEBUG"`)

	viper.SetDefault("port", 6041)
	_ = viper.BindEnv("port", "BLM_PORT")
	pflag.IntP("port", "P", 6041, `http port. Env "BLM_PORT"`)

	viper.SetDefault("logLevel", "info")
	_ = viper.BindEnv("logLevel", "BLM_LOG_LEVEL")
	pflag.String("logLevel", "info", `log level (panic fatal error warn warning info debug trace). Env "BLM_LOG_LEVEL"`)

	initLog()
	initSSL()
	initCors()
	initPool()

	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		panic(err)
	}
}
