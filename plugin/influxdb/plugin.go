package influxdb

import (
	"bufio"
	"fmt"
	"github.com/gin-gonic/gin"
	dbPackage "github.com/taosdata/blm3/db"
	"github.com/taosdata/blm3/log"
	"github.com/taosdata/blm3/plugin"
	"github.com/taosdata/blm3/tools"
	"github.com/taosdata/blm3/tools/pool"
	"github.com/taosdata/blm3/tools/web"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/taosdata/driver-go/v2/af"
	"io"
	"net/http"
	"strings"
	"time"
)

var logger = log.GetLogger("influxdb")

type Influxdb struct {
	conf        Config
	reserveConn *af.Connector
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
	if p.reserveConn != nil {
		p.reserveConn.Close()
	}
	return nil
}

func (p *Influxdb) write(c *gin.Context) {
	id := web.GetRequestID(c)
	logger := logger.WithField("sessionID", id)
	var lines []string
	rd := bufio.NewReader(c.Request.Body)
	precision := c.Query("precision")
	if len(precision) == 0 {
		precision = "ns"
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
			lines = append(lines, tmp.String())
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
	taosConn, err := dbPackage.GetAdvanceConnection(user, password)
	if err != nil {
		logger.WithError(err).Errorln("connect taosd error")
		p.commonResponse(c, http.StatusInternalServerError, &message{Code: "internal error", Message: err.Error()})
		return
	}
	defer taosConn.Put()
	conn := taosConn.TaosConnection
	_, err = conn.Exec(fmt.Sprintf("create database if not exists %s precision 'ns'", db))
	if err != nil {
		logger.WithError(err).Errorln("create database error", db)
		p.commonResponse(c, http.StatusInternalServerError, &message{Code: "internal error", Message: err.Error()})
		return
	}
	_, err = conn.Exec(fmt.Sprintf("use %s", db))
	if err != nil {
		logger.WithError(err).Error("change to database error", db)
		p.commonResponse(c, http.StatusInternalServerError, &message{Code: "internal error", Message: err.Error()})
		return
	}
	start := time.Now()
	logger.Debugln(start, "insert lines", lines)
	err = conn.InfluxDBInsertLines(lines, precision)
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

func init() {
	_ = viper.BindEnv("influxdb.enable", "BLM_INFLUXDB_ENABLE")
	pflag.Bool("influxdb.enable", true, `enable influxdb. Env "BLM_INFLUXDB_ENABLE"`)
	viper.SetDefault("influxdb.enable", true)
	plugin.Register(&Influxdb{
		conf: Config{},
	})
}
