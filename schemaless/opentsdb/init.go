package opentsdb

import (
	"github.com/taosdata/blm3/log"
)

type Result struct {
	SuccessCount int
	FailCount    int
	ErrorList    []error
}

var logger = log.GetLogger("schemaless").WithField("protocol", "opentsdb")

const valueField = "value"
