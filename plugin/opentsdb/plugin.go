package opentsdb

import (
	"bufio"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/taosdata/blm3/db/commonpool"
	"github.com/taosdata/blm3/schemaless/opentsdb"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/taosdata/blm3/log"
	"github.com/taosdata/blm3/plugin"
	"github.com/taosdata/blm3/tools/pool"
	"github.com/taosdata/blm3/tools/web"
	"github.com/taosdata/driver-go/v2/af"
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
	p.conf.setValue()
	if !p.conf.Enable {
		logger.Info("opentsdb disabled")
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
	isDebug := logger.Logger.IsLevelEnabled(logrus.DebugLevel)
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
	taosConn, err := commonpool.GetConnection(user, password)
	if err != nil {
		logger.WithError(err).Error("connect taosd error")
		p.errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	defer func() {
		putErr := taosConn.Put()
		if putErr != nil {
			logger.WithError(putErr).Errorln("taos connect pool put error")
		}
	}()
	var start time.Time
	if isDebug {
		start = time.Now()
	}
	logger.Debug(start, "insert json payload", string(data))
	err = opentsdb.InsertJson(taosConn.TaosConnection, data, db)
	logger.Debug("insert json payload cost:", time.Now().Sub(start))
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
	isDebug := logger.Logger.IsLevelEnabled(logrus.DebugLevel)
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
	taosConn, err := commonpool.GetConnection(user, password)
	if err != nil {
		logger.WithError(err).Error("connect taosd error")
		p.errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	defer func() {
		putErr := taosConn.Put()
		if putErr != nil {
			logger.WithError(putErr).Errorln("taos connect pool put error")
		}
	}()
	var start time.Time
	if isDebug {
		start = time.Now()
	}
	logger.Debug(start, "insert telnet payload", lines)
	var errorList = make([]error, len(lines))
	for _, line := range lines {
		err := opentsdb.InsertTelnet(taosConn.TaosConnection, line, db)
		if err != nil {
			errorList = append(errorList, err)
		}
	}
	logger.Debug("insert telnet payload cost:", time.Now().Sub(start))
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
	plugin.Register(&Plugin{})
}
