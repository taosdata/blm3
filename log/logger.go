package log

import (
	"fmt"
	"github.com/huskar-t/blm_demo/config"
	"github.com/huskar-t/blm_demo/tools/pool"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"math/rand"
	"path"
	"time"
)

var logger = logrus.New()
var ServerID = randomID()

func ConfigLog() {
	l, err := logrus.ParseLevel(config.Conf.LogLevel)
	if err != nil {
		panic(err)
	}
	logger.SetLevel(l)
	writer, err := rotatelogs.New(
		path.Join(config.Conf.Log.Path, "blm_%Y_%m_%d_%H_%M.log"),
		rotatelogs.WithRotationCount(config.Conf.Log.RotationCount),
		rotatelogs.WithRotationTime(config.Conf.Log.RotationTime),
	)
	if err != nil {
		panic(err)
	}
	logger.SetOutput(writer)
}

func GetLogger(model string) *logrus.Entry {
	return logger.WithFields(logrus.Fields{"model": model})
}
func init() {
	logger.SetFormatter(&TaosLogFormatter{})
}

func randomID() string {
	return fmt.Sprintf("%08v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(100000000))
}

type TaosLogFormatter struct {
}

func (t *TaosLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := pool.BytesPoolGet()
	defer pool.BytesPoolPut(b)
	b.WriteString(entry.Time.Format("01/02 15:04:05.000000"))
	b.WriteByte(' ')
	b.WriteString(ServerID)
	b.WriteString(" BLM ")
	b.WriteString(entry.Level.String())
	b.WriteString(` "`)
	b.WriteString(entry.Message)
	b.WriteByte('"')
	for k, v := range entry.Data {
		b.WriteByte(' ')
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(fmt.Sprintf("%v", v))
	}
	b.WriteByte('\n')
	return b.Bytes(), nil
}
