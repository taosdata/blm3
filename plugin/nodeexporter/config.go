package nodeexporter

import (
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/taosdata/driver-go/v2/common"
)

type Config struct {
	Enable                bool
	DB                    string
	User                  string
	Password              string
	URLs                  []string
	ResponseTimeout       time.Duration
	HttpUsername          string
	HttpPassword          string
	HttpBearerTokenString string
	CaCertFile            string
	CertFile              string
	KeyFile               string
	InsecureSkipVerify    bool
	GatherDuration        time.Duration
}

func (c *Config) setValue() {
	c.Enable = viper.GetBool("node_exporter.enable")
	c.DB = viper.GetString("node_exporter.db")
	c.User = viper.GetString("node_exporter.user")
	c.Password = viper.GetString("node_exporter.password")
	c.URLs = viper.GetStringSlice("node_exporter.urls")
	c.ResponseTimeout = viper.GetDuration("node_exporter.responseTimeout")
	c.HttpUsername = viper.GetString("node_exporter.httpUsername")
	c.HttpPassword = viper.GetString("node_exporter.httpPassword")
	c.HttpBearerTokenString = viper.GetString("node_exporter.httpBearerTokenString")
	c.CaCertFile = viper.GetString("node_exporter.caCertFile")
	c.CertFile = viper.GetString("node_exporter.certFile")
	c.KeyFile = viper.GetString("node_exporter.keyFile")
	c.InsecureSkipVerify = viper.GetBool("node_exporter.insecureSkipVerify")
	c.GatherDuration = viper.GetDuration("node_exporter.gatherDuration")

}

func init() {
	_ = viper.BindEnv("node_exporter.enable", "BLM_NODE_EXPORTER_ENABLE")
	pflag.Bool("node_exporter.enable", false, `enable node_exporter. Env "BLM_NODE_EXPORTER_ENABLE"`)
	viper.SetDefault("node_exporter.enable", false)

	_ = viper.BindEnv("node_exporter.db", "BLM_NODE_EXPORTER_DB")
	pflag.String("node_exporter.db", "node_exporter", `node_exporter db name. Env "BLM_NODE_EXPORTER_DB"`)
	viper.SetDefault("node_exporter.db", "node_exporter")

	_ = viper.BindEnv("node_exporter.user", "BLM_NODE_EXPORTER_USER")
	pflag.String("node_exporter.user", common.DefaultUser, `node_exporter user. Env "BLM_NODE_EXPORTER_USER"`)
	viper.SetDefault("node_exporter.user", common.DefaultUser)

	_ = viper.BindEnv("node_exporter.password", "BLM_NODE_EXPORTER_PASSWORD")
	pflag.String("node_exporter.password", common.DefaultPassword, `node_exporter password. Env "BLM_NODE_EXPORTER_PASSWORD"`)
	viper.SetDefault("node_exporter.password", common.DefaultPassword)

	_ = viper.BindEnv("node_exporter.urls", "BLM_NODE_EXPORTER_URLS")
	pflag.StringSlice("node_exporter.urls", []string{}, `node_exporter urls. Env "BLM_NODE_EXPORTER_URLS"`)
	viper.SetDefault("node_exporter.urls", []string{})

	_ = viper.BindEnv("node_exporter.responseTimeout", "BLM_NODE_EXPORTER_RESPONSE_TIMEOUT")
	pflag.Duration("node_exporter.responseTimeout", 5*time.Second, `node_exporter response timeout. Env "BLM_NODE_EXPORTER_RESPONSE_TIMEOUT"`)
	viper.SetDefault("node_exporter.responseTimeout", "5s")

	_ = viper.BindEnv("node_exporter.httpUsername", "BLM_NODE_EXPORTER_HTTP_USERNAME")
	pflag.String("node_exporter.httpUsername", "", `node_exporter http username. Env "BLM_NODE_EXPORTER_HTTP_USERNAME"`)
	viper.SetDefault("node_exporter.httpUsername", "")

	_ = viper.BindEnv("node_exporter.httpPassword", "BLM_NODE_EXPORTER_HTTP_PASSWORD")
	pflag.String("node_exporter.httpPassword", "", `node_exporter http password. Env "BLM_NODE_EXPORTER_HTTP_PASSWORD"`)
	viper.SetDefault("node_exporter.httpPassword", "")

	_ = viper.BindEnv("node_exporter.httpBearerTokenString", "BLM_NODE_EXPORTER_HTTP_BEARER_TOKEN_STRING")
	pflag.String("node_exporter.httpBearerTokenString", "", `node_exporter http bearer token. Env "BLM_NODE_EXPORTER_HTTP_BEARER_TOKEN_STRING"`)
	viper.SetDefault("node_exporter.httpBearerTokenString", "")

	_ = viper.BindEnv("node_exporter.caCertFile", "BLM_NODE_EXPORTER_CA_CERT_FILE")
	pflag.String("node_exporter.caCertFile", "", `node_exporter ca cert file path. Env "BLM_NODE_EXPORTER_CA_CERT_FILE"`)
	viper.SetDefault("node_exporter.caCertFile", "")

	_ = viper.BindEnv("node_exporter.certFile", "BLM_NODE_EXPORTER_CERT_FILE")
	pflag.String("node_exporter.certFile", "", `node_exporter cert file path. Env "BLM_NODE_EXPORTER_CERT_FILE"`)
	viper.SetDefault("node_exporter.certFile", "")

	_ = viper.BindEnv("node_exporter.keyFile", "BLM_NODE_EXPORTER_KEY_FILE")
	pflag.String("node_exporter.keyFile", "", `node_exporter cert key file path. Env "BLM_NODE_EXPORTER_KEY_FILE"`)
	viper.SetDefault("node_exporter.keyFile", "")

	_ = viper.BindEnv("node_exporter.insecureSkipVerify", "BLM_NODE_EXPORTER_INSECURE_SKIP_VERIFY")
	pflag.Bool("node_exporter.insecureSkipVerify", true, `node_exporter skip ssl check. Env "BLM_NODE_EXPORTER_INSECURE_SKIP_VERIFY"`)
	viper.SetDefault("node_exporter.insecureSkipVerify", true)

	_ = viper.BindEnv("node_exporter.gatherDuration", "BLM_NODE_EXPORTER_GATHER_DURATION")
	pflag.Duration("node_exporter.gatherDuration", 5*time.Second, `node_exporter gather duration. Env "BLM_NODE_EXPORTER_GATHER_DURATION"`)
	viper.SetDefault("node_exporter.gatherDuration", "5s")
}
