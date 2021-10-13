package config

import (
	"github.com/gin-contrib/cors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type CorsConfig struct {
	AllowAllOrigins  bool
	AllowOrigins     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	AllowWebSockets  bool
}

func (conf *CorsConfig) GetConfig() cors.Config {
	corsConfig := cors.DefaultConfig()
	if conf.AllowAllOrigins {
		corsConfig.AllowAllOrigins = true
	} else {
		if len(conf.AllowOrigins) == 0 {
			corsConfig.AllowOrigins = []string{
				"http://127.0.0.1",
			}
		} else {
			corsConfig.AllowOrigins = conf.AllowOrigins
		}
	}
	if len(conf.AllowHeaders) > 0 {
		corsConfig.AddAllowHeaders(conf.AllowHeaders...)
	}

	corsConfig.AllowCredentials = conf.AllowCredentials
	corsConfig.AllowWebSockets = conf.AllowWebSockets
	corsConfig.AllowWildcard = true
	corsConfig.ExposeHeaders = []string{"Authorization"}
	return corsConfig
}

func initCors() {
	viper.SetDefault("cors.allowAllOrigins", false)
	_ = viper.BindEnv("cors.allowAllOrigins", "BLM_CORS_ALLOW_ALL_ORIGINS")
	pflag.Bool("cors.allowAllOrigins", false, `cors allow all origins. Env "BLM_CORS_ALLOW_ALL_ORIGINS"`)

	viper.SetDefault("cors.allowOrigins", nil)
	_ = viper.BindEnv("cors.allowOrigins", "BLM_ALLOW_ORIGINS")
	pflag.StringArray("cors.allowOrigins", nil, `cors allow origins. Env "BLM_ALLOW_ORIGINS"`)

	viper.SetDefault("cors.allowHeaders", nil)
	_ = viper.BindEnv("cors.allowHeaders", "BLM_ALLOW_HEADERS")
	pflag.StringArray("cors.allowHeaders", nil, `cors allow HEADERS. Env "BLM_ALLOW_HEADERS"`)

	viper.SetDefault("cors.exposeHeaders", nil)
	_ = viper.BindEnv("cors.exposeHeaders", "BLM_Expose_Headers")
	pflag.StringArray("cors.exposeHeaders", nil, `cors expose headers. Env "BLM_Expose_Headers"`)

	viper.SetDefault("cors.allowCredentials", false)
	_ = viper.BindEnv("cors.allowCredentials", "BLM_CORS_ALLOW_Credentials")
	pflag.Bool("cors.allowCredentials", false, `cors allow credentials. Env "BLM_CORS_ALLOW_Credentials"`)

	viper.SetDefault("cors.allowWebSockets", false)
	_ = viper.BindEnv("cors.allowWebSockets", "BLM_CORS_ALLOW_WebSockets")
	pflag.Bool("cors.allowWebSockets", false, `cors allow WebSockets. Env "BLM_CORS_ALLOW_WebSockets"`)

}

func (conf *CorsConfig) setValue() {
	conf.AllowAllOrigins = viper.GetBool("cors.allowAllOrigins")
	conf.AllowOrigins = viper.GetStringSlice("cors.allowOrigins")
	conf.AllowHeaders = viper.GetStringSlice("cors.allowHeaders")
	conf.ExposeHeaders = viper.GetStringSlice("cors.exposeHeaders")
	conf.AllowCredentials = viper.GetBool("cors.allowCredentials")
	conf.AllowWebSockets = viper.GetBool("cors.allowWebSockets")
}
