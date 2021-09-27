package config

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/taosdata/go-utils/util"
	"github.com/taosdata/go-utils/web"
	"io/ioutil"
	"log"
)

type Config struct {
	Cors     web.CorsConfig
	Debug    bool
	Port     int
	LogLevel string
	P        toml.Primitive
}

var (
	Conf        *Config
	configPath  = "./config/blm3.toml"
	configBytes []byte
)

func init() {
	cp := flag.String("config path", "", "default ./config/blm3.toml")
	flag.Parse()
	if *cp != "" {
		configPath = *cp
	}
	fmt.Println("load config :", configPath)
	var conf Config
	var err error
	if util.PathExist(configPath) {
		configBytes, err = ioutil.ReadFile(configPath)
		if err != nil {
			log.Fatal(err)
		}
		_, err = toml.Decode(string(configBytes), &conf)
		if err != nil {
			log.Fatal(err)
		}
	}
	conf.Cors.Init()
	if conf.Port == 0 {
		conf.Port = 6041
	}
	if conf.LogLevel == "" {
		conf.LogLevel = "info"
	}
	Conf = &conf
}

func Decode(v interface{}) error {
	_, err := toml.Decode(string(configBytes), v)
	return err
}

func Clear() {
	configBytes = nil
}
