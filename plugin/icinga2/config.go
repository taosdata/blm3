package icinga2

import (
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/taosdata/driver-go/v2/common"
)

type Config struct {
	Enable             bool
	DB                 string
	User               string
	Password           string
	Host               string
	ResponseTimeout    time.Duration
	HttpUsername       string
	HttpPassword       string
	CaCertFile         string
	CertFile           string
	KeyFile            string
	InsecureSkipVerify bool
	GatherDuration     time.Duration
}

func (c *Config) setValue() {
	c.Enable = viper.GetBool("icinga2.enable")
	c.DB = viper.GetString("icinga2.db")
	c.User = viper.GetString("icinga2.user")
	c.Password = viper.GetString("icinga2.password")
	c.Host = viper.GetString("icinga2.host")
	c.ResponseTimeout = viper.GetDuration("icinga2.responseTimeout")
	c.HttpUsername = viper.GetString("icinga2.httpUsername")
	c.HttpPassword = viper.GetString("icinga2.httpPassword")
	c.CaCertFile = viper.GetString("icinga2.caCertFile")
	c.CertFile = viper.GetString("icinga2.certFile")
	c.KeyFile = viper.GetString("icinga2.keyFile")
	c.InsecureSkipVerify = viper.GetBool("icinga2.insecureSkipVerify")
	c.GatherDuration = viper.GetDuration("icinga2.gatherDuration")
}

func init() {
	_ = viper.BindEnv("icinga2.enable", "BLM_ICINGA2_ENABLE")
	pflag.Bool("icinga2.enable", false, `enable icinga2. Env "BLM_ICINGA2_ENABLE"`)
	viper.SetDefault("icinga2.enable", false)

	_ = viper.BindEnv("icinga2.db", "BLM_ICINGA2_DB")
	pflag.String("icinga2.db", "icinga2", `icinga2 db name. Env "BLM_ICINGA2_DB"`)
	viper.SetDefault("icinga2.db", "icinga2")

	_ = viper.BindEnv("icinga2.user", "BLM_ICINGA2_USER")
	pflag.String("icinga2.user", common.DefaultUser, `icinga2 user. Env "BLM_ICINGA2_USER"`)
	viper.SetDefault("icinga2.user", common.DefaultUser)

	_ = viper.BindEnv("icinga2.password", "BLM_ICINGA2_PASSWORD")
	pflag.String("icinga2.password", common.DefaultPassword, `icinga2 password. Env "BLM_ICINGA2_PASSWORD"`)
	viper.SetDefault("icinga2.password", common.DefaultPassword)

	_ = viper.BindEnv("icinga2.host", "BLM_ICINGA2_HOST")
	pflag.String("icinga2.host", "", `icinga2 server restful host. Env "BLM_ICINGA2_HOST"`)
	viper.SetDefault("icinga2.host", "")

	_ = viper.BindEnv("icinga2.responseTimeout", "BLM_ICINGA2_RESPONSE_TIMEOUT")
	pflag.Duration("icinga2.responseTimeout", 5*time.Second, `icinga2 response timeout. Env "BLM_ICINGA2_RESPONSE_TIMEOUT"`)
	viper.SetDefault("icinga2.responseTimeout", "5s")

	_ = viper.BindEnv("icinga2.httpUsername", "BLM_ICINGA2_HTTP_USERNAME")
	pflag.String("icinga2.httpUsername", "", `icinga2 http username. Env "BLM_ICINGA2_HTTP_USERNAME"`)
	viper.SetDefault("icinga2.httpUsername", "")

	_ = viper.BindEnv("icinga2.httpPassword", "BLM_ICINGA2_HTTP_PASSWORD")
	pflag.String("icinga2.httpPassword", "", `icinga2 http password. Env "BLM_ICINGA2_HTTP_PASSWORD"`)
	viper.SetDefault("icinga2.httpPassword", "")

	_ = viper.BindEnv("icinga2.caCertFile", "BLM_ICINGA2_CA_CERT_FILE")
	pflag.String("icinga2.caCertFile", "", `icinga2 ca cert file path. Env "BLM_ICINGA2_CA_CERT_FILE"`)
	viper.SetDefault("icinga2.caCertFile", "")

	_ = viper.BindEnv("icinga2.certFile", "BLM_ICINGA2_CERT_FILE")
	pflag.String("icinga2.certFile", "", `icinga2 cert file path. Env "BLM_ICINGA2_CERT_FILE"`)
	viper.SetDefault("icinga2.certFile", "")

	_ = viper.BindEnv("icinga2.keyFile", "BLM_ICINGA2_KEY_FILE")
	pflag.String("icinga2.keyFile", "", `icinga2 cert key file path. Env "BLM_ICINGA2_KEY_FILE"`)
	viper.SetDefault("icinga2.keyFile", "")

	_ = viper.BindEnv("icinga2.insecureSkipVerify", "BLM_ICINGA2_INSECURE_SKIP_VERIFY")
	pflag.Bool("icinga2.insecureSkipVerify", true, `icinga2 skip ssl check. Env "BLM_ICINGA2_INSECURE_SKIP_VERIFY"`)
	viper.SetDefault("icinga2.insecureSkipVerify", true)

	_ = viper.BindEnv("icinga2.gatherDuration", "BLM_ICINGA2_GATHER_DURATION")
	pflag.Duration("icinga2.gatherDuration", 5*time.Second, `icinga2 gather duration. Env "BLM_ICINGA2_GATHER_DURATION"`)
	viper.SetDefault("icinga2.gatherDuration", "5s")
}
