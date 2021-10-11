package rest

import (
	"database/sql/driver"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/huskar-t/blm_demo/httperror"
	"github.com/huskar-t/blm_demo/log"
	"github.com/huskar-t/blm_demo/tools/web"
	"github.com/taosdata/driver-go/v2/common"
	tErrors "github.com/taosdata/driver-go/v2/errors"
	"github.com/taosdata/driver-go/v2/wrapper"
	"net/http"
	"strings"
	"time"
	"unsafe"
)

const LayoutMillSecond = "2006-01-02 15:04:05.000"
const LayoutMicroSecond = "2006-01-02 15:04:05.000000"
const LayoutNanoSecond = "2006-01-02 15:04:05.000000000"

var logger = log.GetLogger("restful")

type Restful struct {
}

func (ctl *Restful) Init(r gin.IRouter) error {
	api := r.Group("rest")
	api.POST(":name", checkAuth, ctl.sql)
	api.GET("login/:user/:password", ctl.des)
	return nil
}

func (ctl *Restful) sql(c *gin.Context) {
	n := c.Param("name")
	if len(n) == 0 {
		errorResponse(c, httperror.HTTP_UNSUPPORT_URL)
		return
	}
	switch n {
	case "sql":
		doQuery(c, func(ts int64, precision int) driver.Value {
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
	case "sqlt":
		doQuery(c, func(ts int64, precision int) driver.Value {
			return ts
		})
	case "sqlutc":
		doQuery(c, func(ts int64, precision int) driver.Value {
			return common.TimestampConvertToTime(ts, precision).Format(time.RFC3339Nano)
		})
	default:
		errorResponse(c, httperror.HTTP_UNSUPPORT_URL)
		return
	}
}

type TDEngineRestfulResp struct {
	Status     string           `json:"status"`
	Head       []string         `json:"head"`
	Data       [][]driver.Value `json:"data"`
	ColumnMeta [][]interface{}  `json:"column_meta"`
	Rows       int              `json:"rows"`
}

func doQuery(c *gin.Context, timeFunc wrapper.FormatTimeFunc) {
	var taos unsafe.Pointer
	var result unsafe.Pointer
	id := web.GetRequestID(c)
	logger := logger.WithField("sessionID", id)
	defer func() {
		err := recover()
		if err != nil {
			logger.Errorln(err)
		}
		if result != nil {
			wrapper.TaosFreeResult(result)
		}
		if taos != nil {
			wrapper.TaosClose(taos)
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
	taos, err = wrapper.TaosConnect("", user, password, "", 0)
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
	start := time.Now()
	logger.Debugln(start, "start execute sql:", sql)
	result = wrapper.TaosQuery(taos, sql)
	logger.Debugln("execute sql finish cast:", time.Now().Sub(start))
	code := wrapper.TaosError(result)
	if code != httperror.SUCCESS {
		errorMsg := wrapper.TaosErrorStr(result)
		logger.Errorln("execute sql error:", sql, code&0xffff, errorMsg)
		errorResponseWithMsg(c, int(code), errorMsg)
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
		header, _ := wrapper.ReadColumn(result, numFields)
		var data = make([][]driver.Value, 0)
		for {
			blockSize, block := wrapper.TaosFetchBlock(result)
			if blockSize == 0 {
				break
			}
			d := wrapper.ReadBlockWithTimeFormat(result, block, blockSize, header.ColLength, header.ColTypes, timeFunc)
			data = append(data, d...)
		}
		logger.Debugln("execute sql success return data rows:", len(data))
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
