package db

import (
	"sync"

	"github.com/taosdata/blm3/config"
	"github.com/taosdata/blm3/db/async"
	"github.com/taosdata/blm3/log"
	"github.com/taosdata/driver-go/v2/common"
	"github.com/taosdata/driver-go/v2/errors"
	"github.com/taosdata/driver-go/v2/wrapper"
)

var once = sync.Once{}
var logger = log.GetLogger("db")

func PrepareConnection() {
	if len(config.Conf.TaosConfigDir) != 0 {
		once.Do(func() {
			code := wrapper.TaosOptions(common.TSDB_OPTION_CONFIGDIR, config.Conf.TaosConfigDir)
			err := errors.GetError(code)
			if err != nil {
				logger.WithError(err).Panic("set taos config file ", config.Conf.TaosConfigDir)
			}
		})
	}
	async.GlobalAsync = async.NewAsync(async.NewHandlerPool(10000))
}
