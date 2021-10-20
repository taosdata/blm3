package rest

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"strings"
	"time"
	"unsafe"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/taosdata/blm3/db/commonpool"
	"github.com/taosdata/blm3/httperror"
	"github.com/taosdata/blm3/log"
	"github.com/taosdata/blm3/tools/web"
	"github.com/taosdata/driver-go/v2/common"
	tErrors "github.com/taosdata/driver-go/v2/errors"
	"github.com/taosdata/driver-go/v2/wrapper"
)

const LayoutMillSecond = "2006-01-02 15:04:05.000"
const LayoutMicroSecond = "2006-01-02 15:04:05.000000"
const LayoutNanoSecond = "2006-01-02 15:04:05.000000000"

var logger = log.GetLogger("restful")

type Restful struct {
	reserveConn unsafe.Pointer
}

func (ctl *Restful) Init(r gin.IRouter) error {
	api := r.Group("rest")
	api.POST("sql", checkAuth, ctl.sql)
	api.POST("sqlt", checkAuth, ctl.sqlt)
	api.POST("sqlutc", checkAuth, ctl.sqlutc)
	api.POST("sql/:db", checkAuth, ctl.sql)
	api.POST("sqlt/:db", checkAuth, ctl.sqlt)
	api.POST("sqlutc/:db", checkAuth, ctl.sqlutc)
	api.GET("login/:user/:password", ctl.des)
	return nil
}

func (ctl *Restful) sql(c *gin.Context) {
	db := c.Param("db")
	ctl.doQuery(c, db, func(ts int64, precision int) driver.Value {
		switch precision {
		case common.PrecisionMilliSecond:
			return common.TimestampConvertToTime(ts, precision).Local().Format(LayoutMillSecond)
		case common.PrecisionMicroSecond:
			return common.TimestampConvertToTime(ts, precision).Local().Format(LayoutMicroSecond)
		case common.PrecisionNanoSecond:
			return common.TimestampConvertToTime(ts, precision).Local().Format(LayoutNanoSecond)
		}
		panic("unsupported precision")
	})
}
func (ctl *Restful) sqlt(c *gin.Context) {
	db := c.Param("db")
	ctl.doQuery(c, db, func(ts int64, precision int) driver.Value {
		return ts
	})

}
func (ctl *Restful) sqlutc(c *gin.Context) {
	db := c.Param("db")
	ctl.doQuery(c, db, func(ts int64, precision int) driver.Value {
		return common.TimestampConvertToTime(ts, precision).Format(time.RFC3339Nano)
	})
}

type TDEngineRestfulResp struct {
	Status     string           `json:"status"`
	Head       []string         `json:"head"`
	ColumnMeta [][]interface{}  `json:"column_meta"`
	Data       [][]driver.Value `json:"data"`
	Rows       int              `json:"rows"`
}

func (ctl *Restful) doQuery(c *gin.Context, db string, timeFunc wrapper.FormatTimeFunc) {
	var result unsafe.Pointer
	var s time.Time
	isDebug := logger.Logger.IsLevelEnabled(logrus.DebugLevel)
	id := web.GetRequestID(c)
	logger := logger.WithField("sessionID", id)
	defer func() {
		if result != nil {
			wrapper.TaosFreeResult(result)
		}
	}()
	b, err := c.GetRawData()
	if err != nil {
		logger.WithError(err).Error("get request body error")
		errorResponse(c, httperror.HTTP_INVALID_CONTENT_LENGTH)
		return
	}
	if len(b) == 0 {
		logger.Errorln("no msg got")
		errorResponse(c, httperror.HTTP_NO_MSG_INPUT)
		return
	}
	sql := strings.TrimSpace(string(b))
	if len(sql) == 0 {
		logger.Errorln("no sql got")
		errorResponse(c, httperror.HTTP_NO_SQL_INPUT)
		return
	}
	user := c.MustGet(UserKey).(string)
	password := c.MustGet(PasswordKey).(string)
	if isDebug {
		s = time.Now()
	}
	taosConnect, err := commonpool.GetConnection(user, password)

	logger.Debugln("taos connect cost:", time.Now().Sub(s))
	if err != nil {
		logger.WithError(err).Error("connect taosd error")
		var tError *tErrors.TaosError
		if errors.As(err, &tError) {
			errorResponseWithMsg(c, int(tError.Code), tError.ErrStr)
			return
		} else {
			errorResponseWithMsg(c, 0xffff, err.Error())
			return
		}
	}
	defer func() {
		if isDebug {
			s = time.Now()
		}
		err := taosConnect.Put()
		if err != nil {
			panic(err)
		}
		logger.Debugln("taos put connect cost:", time.Now().Sub(s))
	}()
	if len(db) > 0 {
		if isDebug {
			s = time.Now()
		}
		code := wrapper.TaosSelectDB(taosConnect.TaosConnection, db)
		logger.Debugln("taos select db cost:", time.Now().Sub(s))
		if code != httperror.SUCCESS {
			if isDebug {
				s = time.Now()
			}
			errorMsg := wrapper.TaosErrorStr(result)
			logger.Debugln("taos select db get error string cost:", time.Now().Sub(s))
			logger.Errorln("taos select db error:", sql, code&0xffff, errorMsg)
			errorResponseWithMsg(c, code, errorMsg)
			return
		}
	}
	startExec := time.Now()
	logger.Debugln(startExec, "start execute sql:", sql)
	result = wrapper.TaosQuery(taosConnect.TaosConnection, sql)
	logger.Debugln("execute sql cost:", time.Now().Sub(startExec))
	if isDebug {
		s = time.Now()
	}
	code := wrapper.TaosError(result)
	logger.Debugln("taos get error cost:", time.Now().Sub(s))
	if code != httperror.SUCCESS {
		if isDebug {
			s = time.Now()
		}
		errorMsg := wrapper.TaosErrorStr(result)
		logger.Debugln("taos get error string cost:", time.Now().Sub(s))
		logger.Errorln("taos execute sql error:", sql, code&0xffff, errorMsg)
		errorResponseWithMsg(c, code, errorMsg)
		return
	}
	numFields := wrapper.TaosFieldCount(result)
	if numFields == 0 {
		// there are no select and show kinds of commands
		affectedRows := wrapper.TaosAffectedRows(result)
		logger.Debugln("execute sql success affected rows:", affectedRows)
		c.JSON(http.StatusOK, &TDEngineRestfulResp{
			Status:     "succ",
			Head:       []string{"affected_rows"},
			Data:       [][]driver.Value{{affectedRows}},
			ColumnMeta: [][]interface{}{{"affected_rows", 4, 4}},
			Rows:       1,
		})
	} else {
		if isDebug {
			s = time.Now()
		}
		header, _ := wrapper.ReadColumn(result, numFields)
		logger.Debugln("taos read column cost:", time.Now().Sub(s))
		var data = make([][]driver.Value, 0)
		if isDebug {
			s = time.Now()
		}
		for {
			blockSize, block := wrapper.TaosFetchBlock(result)
			if blockSize == 0 {
				break
			}
			d := wrapper.ReadBlockWithTimeFormat(result, block, blockSize, header.ColLength, header.ColTypes, timeFunc)
			data = append(data, d...)
		}
		logger.Debugln("execute sql success return data rows:", len(data), ",cost:", time.Now().Sub(s))
		var columnMeta [][]interface{}
		for i := 0; i < len(header.ColNames); i++ {
			columnMeta = append(columnMeta, []interface{}{
				header.ColNames[i],
				header.ColTypes[i],
				header.ColLength[i],
			})
		}
		c.JSON(http.StatusOK, &TDEngineRestfulResp{
			Status:     "succ",
			Head:       header.ColNames,
			Data:       data,
			ColumnMeta: columnMeta,
			Rows:       len(data),
		})
	}
}

func (ctl *Restful) des(c *gin.Context) {
	user := c.Param("user")
	password := c.Param("password")
	if len(user) < 0 || len(user) > 24 || len(password) < 0 || len(password) > 24 {
		errorResponse(c, httperror.HTTP_GEN_TAOSD_TOKEN_ERR)
		return
	}
	//conn, err := commonpool.GetConnection(user, password)
	//if err != nil {
	//	errorResponse(c, httperror.TSDB_CODE_RPC_AUTH_FAILURE)
	//	return
	//}
	//conn.Put()
	token, err := EncodeDes(user, password)
	if err != nil {
		errorResponse(c, httperror.HTTP_GEN_TAOSD_TOKEN_ERR)
		return
	}
	c.JSON(http.StatusOK, &Message{
		Status: "succ",
		Code:   0,
		Desc:   token,
	})
}

func (ctl *Restful) Close() {
	if ctl.reserveConn != nil {
		wrapper.TaosClose(ctl.reserveConn)
	}
}
