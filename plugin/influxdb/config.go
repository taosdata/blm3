package influxdb

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Enable bool
}

func (c *Config) setValue() {
	c.Enable = viper.GetBool("influxdb.enable")
}

func init() {
	_ = viper.BindEnv("influxdb.enable", "BLM_INFLUXDB_ENABLE")
	pflag.Bool("influxdb.enable", true, `enable influxdb. Env "BLM_INFLUXDB_ENABLE"`)
	viper.SetDefault("influxdb.enable", true)
}
