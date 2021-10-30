package log

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"github.com/taosdata/blm3/config"
)

var logger = logrus.New()
var ServerID = randomID()
var globalLogFormatter = &TaosLogFormatter{buffer: &bytes.Buffer{}}

type FileHook struct {
	formatter logrus.Formatter
	writer    io.Writer
	buf       *bytes.Buffer
}

func NewFileHook(formatter logrus.Formatter, writer io.Writer) *FileHook {
	return &FileHook{formatter: formatter, writer: writer, buf: &bytes.Buffer{}}
}

func (f *FileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (f *FileHook) Fire(entry *logrus.Entry) error {
	data, err := f.formatter.Format(entry)
	if err != nil {
		return err
	}
	f.buf.Write(data)
	if f.buf.Len() > 1024 {
		_, err = f.writer.Write(f.buf.Bytes())
		f.buf.Reset()
		if err != nil {
			return err
		}
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
		rotatelogs.WithRotationSize(int64(config.Conf.Log.RotationSize)),
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
	return fmt.Sprintf("%08d", os.Getpid())
}

type TaosLogFormatter struct {
	buffer *bytes.Buffer
}

func (t *TaosLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	t.buffer.Reset()
	t.buffer.WriteString(entry.Time.Format("01/02 15:04:05.000000"))
	t.buffer.WriteByte(' ')
	t.buffer.WriteString(ServerID)
	t.buffer.WriteString(" BLM ")
	t.buffer.WriteString(entry.Level.String())
	t.buffer.WriteString(` "`)
	t.buffer.WriteString(entry.Message)
	t.buffer.WriteByte('"')
	for k, v := range entry.Data {
		t.buffer.WriteByte(' ')
		t.buffer.WriteString(k)
		t.buffer.WriteByte('=')
		t.buffer.WriteString(fmt.Sprintf("%v", v))
	}
	t.buffer.WriteByte('\n')
	return t.buffer.Bytes(), nil
}
