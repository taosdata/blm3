package opentsdb

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	dbPackage "github.com/huskar-t/blm_demo/db"
	"github.com/huskar-t/blm_demo/log"
	"github.com/huskar-t/blm_demo/plugin"
	"github.com/huskar-t/blm_demo/tools/pool"
	"github.com/huskar-t/blm_demo/tools/web"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/taosdata/driver-go/v2/af"
	"io"
	"net/http"
	"time"
)

var logger = log.GetLogger("opentsdb")

type Plugin struct {
	conf        Config
	reserveConn *af.Connector
}

func (p *Plugin) String() string {
	return "opentsdb"
}

func (p *Plugin) Version() string {
	return "v1"
}

func (p *Plugin) Init(r gin.IRouter) error {
	enable := viper.GetBool("opentsdb.enable")
	p.conf = Config{Enable: enable}
	if !p.conf.Enable {
		logger.Info("opentsdb Disabled")
		return nil
	}
	r.POST("put/json/:db", plugin.Auth(p.errorResponse), p.insertJson)
	r.POST("put/telnet/:db", plugin.Auth(p.errorResponse), p.insertTelnet)
	return nil
}

func (p *Plugin) Start() error {
	if !p.conf.Enable {
		return nil
	}
	return nil
}

func (p *Plugin) Stop() error {
	return nil
}

func (p *Plugin) insertJson(c *gin.Context) {
	id := web.GetRequestID(c)
	logger := logger.WithField("sessionID", id)
	db := c.Param("db")
	if len(db) == 0 {
		logger.Errorln("db required")
		p.errorResponse(c, http.StatusBadRequest, errors.New("db required"))
		return
	}
	data, err := c.GetRawData()
	if err != nil {
		logger.WithError(err).Error("get request body error")
		p.errorResponse(c, http.StatusBadRequest, err)
		return
	}
	user, password, err := plugin.GetAuth(c)
	if err != nil {
		logger.WithError(err).Error("get auth error")
		p.errorResponse(c, http.StatusBadRequest, err)
		return
	}
	taosConn, err := dbPackage.GetAdvanceConnection(user, password)
	if err != nil {
		logger.WithError(err).Error("connect taosd error")
		p.errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	defer taosConn.Put()
	conn := taosConn.TaosConnection
	_, err = conn.Exec(fmt.Sprintf("create database if not exists %s", db))
	if err != nil {
		logger.WithError(err).Error("create database error", db)
		p.errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	_, err = conn.Exec(fmt.Sprintf("use %s", db))
	if err != nil {
		logger.WithError(err).Error("change to database error", db)
		p.errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	start := time.Now()
	logger.Debug(start, "insert json payload", string(data))
	err = conn.OpenTSDBInsertJsonPayload(string(data))
	logger.Debug("insert json payload cast:", time.Now().Sub(start))
	if err != nil {
		logger.WithError(err).Error("insert json payload error", string(data))
		p.errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	p.successResponse(c)
}

func (p *Plugin) insertTelnet(c *gin.Context) {
	id := web.GetRequestID(c)
	logger := logger.WithField("sessionID", id)
	db := c.Param("db")
	if len(db) == 0 {
		logger.Errorln("db required")
		p.errorResponse(c, http.StatusBadRequest, errors.New("db required"))
		return
	}
	rd := bufio.NewReader(c.Request.Body)
	var lines []string
	tmp := pool.BytesPoolGet()
	defer pool.BytesPoolPut(tmp)
	for {
		l, hasNext, err := rd.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				p.errorResponse(c, http.StatusBadRequest, err)
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
		logger.WithError(err).Error("get auth error")
		p.errorResponse(c, http.StatusBadRequest, err)
		return
	}
	taosConn, err := dbPackage.GetAdvanceConnection(user, password)
	if err != nil {
		logger.WithError(err).Error("connect taosd error")
		p.errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	defer taosConn.Put()
	conn := taosConn.TaosConnection
	_, err = conn.Exec(fmt.Sprintf("create database if not exists %s", db))
	if err != nil {
		logger.WithError(err).Error("create database error", db)
		p.errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	_, err = conn.Exec(fmt.Sprintf("use %s", db))
	if err != nil {
		logger.WithError(err).Error("change to database error", db)
		p.errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	start := time.Now()
	logger.Debug(start, "insert telnet payload", lines)
	err = conn.OpenTSDBInsertTelnetLines(lines)
	logger.Debug("insert telnet payload cast:", time.Now().Sub(start))
	if err != nil {
		logger.WithError(err).Error("insert telnet payload error", lines)
		p.errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	p.successResponse(c)
}

type message struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

func (p *Plugin) errorResponse(c *gin.Context, code int, err error) {
	c.JSON(code, message{
		Code:    code,
		Message: err.Error(),
	})
}

func (p *Plugin) successResponse(c *gin.Context) {
	c.JSON(http.StatusOK, message{Code: http.StatusOK})
}

func init() {
	_ = viper.BindEnv("opentsdb.enable", "BLM_OPENTSDB_ENABLE")
	pflag.Bool("opentsdb.enable", true, `enable opentsdb. Env "BLM_OPENTSDB_ENABLE"`)
	viper.SetDefault("opentsdb.enable", true)
	plugin.Register(&Plugin{
		conf: Config{},
	})
}
