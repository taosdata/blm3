package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Pool struct {
	MaxConnect int
	MaxIdle    int
}

func initPool() {
	viper.SetDefault("pool.maxConnect", 4000)
	_ = viper.BindEnv("pool.maxConnect", "BLM_POOL_MAX_CONNECT")
	pflag.Int("pool.maxConnect", 4000, `max connections to taosd. Env "BLM_POOL_MAX_CONNECT"`)

	viper.SetDefault("pool.maxIdle", 4000)
	_ = viper.BindEnv("pool.maxIdle", "BLM_POOL_MAX_IDLE")
	pflag.Int("pool.maxIdle", 4000, `max idle connections to taosd. Env "BLM_POOL_MAX_IDLE"`)
}

func (p *Pool) setValue() {
	p.MaxConnect = viper.GetInt("pool.maxConnect")
	p.MaxIdle = viper.GetInt("pool.maxIdle")
}
