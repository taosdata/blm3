package log

import (
	"github.com/huskar-t/blm_demo/config"
	"github.com/sirupsen/logrus"
	"github.com/taosdata/go-utils/log"
)

var Logger = log.NewLogger("blm3")

func init() {
	log.SetLevel(config.Conf.LogLevel)
}

func GetLogger(model string) *logrus.Entry {
	return Logger.WithFields(logrus.Fields{"model": model})
}
