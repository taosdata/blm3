package influxdb

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/huskar-t/blm_demo/log"
	"github.com/huskar-t/blm_demo/plugin"
	"github.com/huskar-t/blm_demo/tools"
	"github.com/huskar-t/blm_demo/tools/pool"
	"github.com/huskar-t/blm_demo/tools/web"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/taosdata/driver-go/v2/af"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var logger = log.GetLogger("influxdb")

type Influxdb struct {
	conf Config
}

func (p *Influxdb) String() string {
	return "influxdb"
}

func (p *Influxdb) Version() string {
	return "v1"
}

func (p *Influxdb) Init(r gin.IRouter) error {
	enable := viper.GetBool("influxdb.enable")
	p.conf = Config{Enable: enable}
	if !p.conf.Enable {
		logger.Info("influxdb Disabled")
		return nil
	}
	r.POST("write", getAuth, p.write)
	return nil
}

func (p *Influxdb) Start() error {
	if !p.conf.Enable {
		return nil
	}
	return nil
}

func (p *Influxdb) Stop() error {
	return nil
}

func (p *Influxdb) write(c *gin.Context) {
	id := web.GetRequestID(c)
	logger := logger.WithField("sessionID", id)
	var lines []string
	rd := bufio.NewReader(c.Request.Body)
	precisionReq := c.Query("precision")
	var precision Precision
	var ok bool
	if len(precisionReq) == 0 {
		precision = PrecisionNanoSecond
	} else {
		precision, ok = convertPrecision(precisionReq)
		if !ok {
			logger.Errorln("unknown precision", precisionReq)
			p.badRequestResponse(c, &badRequest{
				Code:    "invalid",
				Message: "unknown precision",
				Op:      "convert precision",
				Err:     fmt.Sprintf("recision %s unknown", precisionReq),
				Line:    0,
			})
			return
		}
	}

	tmp := pool.BytesPoolGet()
	defer pool.BytesPoolPut(tmp)
	for {
		l, hasNext, err := rd.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				logger.Errorln("read line error", err)
				p.badRequestResponse(c, &badRequest{
					Code:    "internal error",
					Message: "read line error",
					Op:      "read line",
					Err:     err.Error(),
					Line:    len(lines),
				})
				return
			}
		}
		tmp.Write(l)
		if !hasNext {
			line, err := convertLine(tmp, precision)
			if err != nil {
				logger.Errorln("convert line error", err)
				p.badRequestResponse(c, &badRequest{
					Code:    "invalid",
					Message: "convert line error",
					Op:      "convert line",
					Err:     err.Error(),
					Line:    len(lines),
				})
				return
			}
			lines = append(lines, line)
			tmp.Reset()
		}
	}
	user, password, err := plugin.GetAuth(c)
	if err != nil {
		p.commonResponse(c, http.StatusUnauthorized, &message{
			Code:    "forbidden",
			Message: err.Error(),
		})
		return
	}
	db := c.Query("db")
	if len(db) == 0 {
		logger.Errorln("db required")
		p.badRequestResponse(c, &badRequest{
			Code:    "not found",
			Message: "db required",
			Op:      "get db query",
			Err:     "db required",
			Line:    0,
		})
		return
	}

	conn, err := af.Open("", user, password, db, 0)
	if err != nil {
		p.commonResponse(c, http.StatusInternalServerError, &message{Code: "internal error", Message: err.Error()})
		return
	}
	defer conn.Close()
	_, err = conn.Exec(fmt.Sprintf("create database if not exist %s precision 'ns'", db))
	if err != nil {
		logger.WithError(err).Errorln("create database error", db)
		p.commonResponse(c, http.StatusInternalServerError, &message{Code: "internal error", Message: err.Error()})
		return
	}
	start := time.Now()
	logger.Debugln(start, "insert lines", lines)
	err = conn.InsertLines(lines)
	logger.Debugln("insert lines finish cast:", time.Now().Sub(start))
	if err != nil {
		logger.WithError(err).Errorln("insert lines error", lines)
		p.commonResponse(c, http.StatusInternalServerError, &message{Code: "internal error", Message: err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

type badRequest struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Op      string `json:"op"`
	Err     string `json:"err"`
	Line    int    `json:"line"`
}
type message struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (p *Influxdb) badRequestResponse(c *gin.Context, resp *badRequest) {
	c.JSON(http.StatusBadRequest, resp)
}
func (p *Influxdb) commonResponse(c *gin.Context, code int, resp *message) {
	c.JSON(code, resp)
}

func getAuth(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	if len(auth) != 0 {
		auth = strings.TrimSpace(auth)
		if strings.HasPrefix(auth, "Basic") {
			user, password, err := tools.DecodeBasic(auth[6:])
			if err == nil {
				c.Set(plugin.UserKey, user)
				c.Set(plugin.PasswordKey, password)
			}
		}
	}

	user := c.Query("u")
	password := c.Query("p")
	if len(user) != 0 {
		c.Set(plugin.UserKey, user)
	}
	if len(password) != 0 {
		c.Set(plugin.PasswordKey, password)
	}
}

type Precision int

//ns - nanoseconds
//u or µ - microseconds
//ms - milliseconds
//s - seconds
//m - minutes
//h - hours
const (
	PrecisionNanoSecond Precision = iota + 1
	PrecisionMicroSecond
	PrecisionMillSecond
	PrecisionSecond
	PrecisionMinute
	PrecisionHour
)

func convertPrecision(precision string) (Precision, bool) {
	switch precision {
	case "ns":
		return PrecisionNanoSecond, true
	case "u", "µ":
		return PrecisionMicroSecond, true
	case "ms":
		return PrecisionMillSecond, true
	case "s":
		return PrecisionSecond, true
	case "m":
		return PrecisionMinute, true
	case "h":
		return PrecisionHour, true
	default:
		return 0, false
	}
}

func convertLine(line *bytes.Buffer, precision Precision) (string, error) {
	switch precision {
	case PrecisionNanoSecond:
		line.WriteString("ns")
		return line.String(), nil
	case PrecisionMicroSecond:
		line.WriteString("us")
		return line.String(), nil
	case PrecisionMillSecond:
		line.WriteString("ms")
		return line.String(), nil
	case PrecisionSecond:
		line.WriteString("s")
		return line.String(), nil
	case PrecisionMinute, PrecisionHour:
		l := line.String()
		splitIndex := 0
		for i := len(l) - 1; i > 0; i-- {
			if l[i] == ' ' {
				splitIndex = i + 1
				break
			}
		}
		if splitIndex == 0 {
			return "", fmt.Errorf("line format error %s", l)
		}
		t, err := strconv.ParseInt(l[splitIndex:], 10, 64)
		if err != nil {
			return "", err
		}
		b := bytes.NewBufferString(l[:splitIndex])
		multiple := int64(1)
		if precision == PrecisionMinute {
			multiple = 60
		} else if precision == PrecisionHour {
			multiple = 3600
		}
		ts := strconv.FormatInt(t*multiple, 10)
		b.WriteString(ts)
		b.WriteString("s")
		return b.String(), nil
	default:
		return "", errors.New("unknown precision")
	}
}

func init() {
	_ = viper.BindEnv("influxdb.enable", "BLM_INFLUXDB_ENABLE")
	pflag.Bool("influxdb.enable", true, `enable influxdb. Env "BLM_INFLUXDB_ENABLE"`)
	viper.SetDefault("influxdb.enable", true)
	plugin.Register(&Influxdb{
		conf: Config{},
	})
}
