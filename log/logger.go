package log

import (
	"fmt"
	"github.com/huskar-t/blm_demo/config"
	"github.com/huskar-t/blm_demo/tools/pool"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"io"
	"math/rand"
	"path"
	"time"
)

var logger = logrus.New()
var ServerID = randomID()
var globalLogFormatter = &TaosLogFormatter{}

type FileHook struct {
	formatter logrus.Formatter
	writer    io.Writer
}

func NewFileHook(formatter logrus.Formatter, writer io.Writer) *FileHook {
	return &FileHook{formatter: formatter, writer: writer}
}

func (f *FileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (f *FileHook) Fire(entry *logrus.Entry) error {
	data, err := f.formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = f.writer.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func ConfigLog() {
	err := SetLevel(config.Conf.LogLevel)
	if err != nil {
		panic(err)
	}
	writer, err := rotatelogs.New(
		path.Join(config.Conf.Log.Path, "blm_%Y_%m_%d_%H_%M.log"),
		rotatelogs.WithRotationCount(config.Conf.Log.RotationCount),
		rotatelogs.WithRotationTime(config.Conf.Log.RotationTime),
	)
	if err != nil {
		panic(err)
	}
	hook := NewFileHook(globalLogFormatter, writer)
	logger.AddHook(hook)
}

func SetLevel(level string) error {
	l, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logger.SetLevel(l)
	return nil
}

func GetLogger(model string) *logrus.Entry {
	return logger.WithFields(logrus.Fields{"model": model})
}
func init() {
	logger.SetFormatter(globalLogFormatter)
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
