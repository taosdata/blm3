package opentsdb

import (
	"bufio"
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/huskar-t/blm_demo/config"
	"github.com/huskar-t/blm_demo/db"
	"github.com/huskar-t/blm_demo/log"
	"github.com/huskar-t/blm_demo/plugin"
	"github.com/taosdata/driver-go/v2/af"
	"io"
	"net/http"
)

var logger = log.GetLogger("opentsdb")

type Plugin struct {
	pool *db.AFConnectorPool
	conf Config
}

func (p *Plugin) String() string {
	return "opentsdb"
}

func (p *Plugin) Version() string {
	return "v1"
}

func (p *Plugin) Init(r gin.IRouter) error {
	err := config.Decode(&p.conf)
	if err != nil {
		return err
	}
	if p.conf.OpenTSDB.Disable {
		logger.Info("opentsdb Disabled")
		return nil
	}
	r.POST("put", plugin.Auth(p.errorResponse), p.insertJson)
	r.POST("put/telnet", plugin.Auth(p.errorResponse), p.insertTelnet)
	return nil
}

func (p *Plugin) Start() error {
	if p.conf.OpenTSDB.Disable {
		return nil
	}
	return nil
}

func (p *Plugin) Stop() error {
	if p.conf.OpenTSDB.Disable {
		return nil
	}
	if p.pool != nil {
		p.pool.Release()
	}
	return nil
}

func (p *Plugin) insertJson(c *gin.Context) {
	data, err := c.GetRawData()
	if err != nil {
		p.errorResponse(c, http.StatusBadRequest, err)
		return
	}
	user, password, err := plugin.GetAuth(c)
	if err != nil {
		p.errorResponse(c, http.StatusBadRequest, err)
		return
	}
	conn, err := af.Open("", user, password, p.conf.OpenTSDB.DB, 0)
	if err != nil {
		p.errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	defer conn.Close()
	err = conn.InsertJsonPayload(string(data))
	if err != nil {
		p.errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	p.successResponse(c)
}

func (p *Plugin) insertTelnet(c *gin.Context) {
	rd := bufio.NewReader(c.Request.Body)
	var lines []string
	tmp := new(bytes.Buffer)
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
			tmp = new(bytes.Buffer)
		}
	}
	user, password, err := plugin.GetAuth(c)
	if err != nil {
		p.errorResponse(c, http.StatusBadRequest, err)
		return
	}
	conn, err := af.Open("", user, password, p.conf.OpenTSDB.DB, 0)
	if err != nil {
		p.errorResponse(c, http.StatusInternalServerError, err)
		return
	}
	defer conn.Close()
	err = conn.InsertTelnetLines(lines)
	if err != nil {
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
	plugin.Register(&Plugin{
		conf: Config{},
	})
}
