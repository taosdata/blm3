package influxdb

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/taosdata/blm3/db/advancedpool"
	"github.com/taosdata/blm3/tools/influxdb/parse"

	"github.com/gin-gonic/gin"
	"github.com/taosdata/blm3/log"
	"github.com/taosdata/blm3/plugin"
	"github.com/taosdata/blm3/tools"
	"github.com/taosdata/blm3/tools/pool"
	"github.com/taosdata/blm3/tools/web"
	"github.com/taosdata/driver-go/v2/af"
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
	p.conf.setValue()
	if !p.conf.Enable {
		logger.Info("influxdb disabled")
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
		return p.reserveConn.Close()
	}
	return nil
}

func min(l int, r int) int {
	if l < r {
		return l
	}
	return r
}

func (p *Influxdb) write(c *gin.Context) {
	id := web.GetRequestID(c)
	logger := logger.WithField("sessionID", id)
	var lines []string
	precision := c.Query("precision")
	if len(precision) == 0 {
		precision = "ns"
	}
	tmp := pool.BytesPoolGet()
	defer pool.BytesPoolPut(tmp)
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
	data, err := c.GetRawData()
	if err != nil {
		logger.WithError(err).Errorln("read line error")
		p.badRequestResponse(c, &badRequest{
			Code:    "internal error",
			Message: "read line error",
			Op:      "read line",
			Err:     err.Error(),
			Line:    0,
		})
		return
	}
	var wrongIndex int
	lines, wrongIndex, err = parse.Repair(data, precision)
	if err != nil {
		logger.WithError(err).Errorln("parse line error")
		p.badRequestResponse(c, &badRequest{
			Code:    "internal error",
			Message: "read line error",
			Op:      "read line",
			Err:     err.Error(),
			Line:    wrongIndex,
		})
		return
	}
	taosConn, err := advancedpool.GetAdvanceConnection(user, password)
	if err != nil {
		logger.WithError(err).Errorln("connect taosd error")
		p.commonResponse(c, http.StatusInternalServerError, &message{Code: "internal error", Message: err.Error()})
		return
	}
	defer func() {
		putErr := taosConn.Put()
		if putErr != nil {
			logger.WithError(putErr).Errorln("taos connect pool put error")
		}
	}()
	conn := taosConn.TaosConnection
	_, err = conn.Exec(fmt.Sprintf("create database if not exists %s precision 'ns' update 2", db))
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
	// logger.Debugln(start, "insert lines", lines)
	succeeded := 0
	batchSize := 10
	for i := 0; i < (len(lines)+batchSize-1)/batchSize; i++ {
		ls := lines[batchSize*i : min(batchSize*(i+1), len(lines))]
		err = conn.InfluxDBInsertLines(ls, precision)
		if err != nil {
			for j := 0; j < len(ls); j++ {
				err = conn.InfluxDBInsertLines(ls[j:j+1], precision)
				if err != nil {
					logger.WithError(err).Errorln("insert line error", ls[j])
					continue
				}
				succeeded += 1
			}
			continue
		}
		succeeded += len(ls)
	}
	if succeeded == 0 {
		p.commonResponse(c, http.StatusInternalServerError, &message{Code: "internal error", Message: err.Error()})
		return
	}
	logger.Debugln("inserted", succeeded, "/", len(lines), " lines finish cost:", time.Since(start))
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
	plugin.Register(&Influxdb{})
}
