package opentsdb

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/taosdata/blm3/schemaless"
	"github.com/taosdata/blm3/tools/pool"
	"sort"
	"strings"
	"time"
	"unicode"
	"unsafe"
)

type PutData struct {
	Metric    string            `json:"metric"`
	Timestamp int64             `json:"timestamp"`
	Value     float64           `json:"value"`
	Tags      map[string]string `json:"tags"`
}

const (
	arrayJson = iota + 1
	objectJson
)

func InsertJson(taosConnect unsafe.Pointer, data []byte, db string) error {
	if len(data) == 0 {
		return fmt.Errorf("empty data")
	}
	executor, err := schemaless.NewExecutor(taosConnect)
	if err != nil {
		return err
	}
	var jsonType = 0
	for _, d := range data {
		if unicode.IsSpace(rune(d)) {
			continue
		} else if d == '{' {
			jsonType = objectJson
			break
		} else if d == '[' {
			jsonType = arrayJson
			break
		} else {
			return fmt.Errorf("unavailable json data")
		}
	}
	var putData []*PutData
	switch jsonType {
	case arrayJson:
		err := json.Unmarshal(data, &putData)
		if err != nil {
			return err
		}
	case objectJson:
		var singleData PutData
		err := json.Unmarshal(data, &singleData)
		if err != nil {
			return err
		}
		putData = []*PutData{&singleData}
	default:
		return fmt.Errorf("unavailable json data")
	}
	b := pool.BytesPoolGet()
	var errList = make([]string, len(putData))
	defer pool.BytesPoolPut(b)
	haveError := false
	for pointIndex, point := range putData {
		b.Reset()
		var ts time.Time
		if point.Timestamp < 10000000000 {
			//second
			ts = time.Unix(point.Timestamp, 0)
		} else {
			//millisecond
			ts = time.Unix(0, point.Timestamp*1e6)
		}
		tagNames := make([]string, len(point.Tags))
		tagValues := make([]string, len(point.Tags))
		index := 0
		for k := range point.Tags {
			tagNames[index] = k
			index += 1
		}
		sort.Strings(tagNames)
		for i, tagName := range tagNames {
			tagValues[i] = point.Tags[tagName]
		}
		b.WriteString(point.Metric)
		for i := 0; i < len(tagNames); i++ {
			b.WriteString(tagNames[i])
			b.WriteByte('=')
			b.WriteString(tagValues[i])
			if i != len(tagNames)-1 {
				b.WriteByte(' ')
			}
		}
		tableName := fmt.Sprintf("_%x", md5.Sum(b.Bytes()))
		b.Reset()
		b.WriteByte('`')
		b.WriteString(point.Metric)
		b.WriteByte('`')
		sql, err := executor.InsertTDengine(&schemaless.InsertLine{
			DB:         db,
			Ts:         ts,
			TableName:  tableName,
			STableName: b.String(),
			Fields: map[string]interface{}{
				valueField: point.Value,
			},
			TagNames:  tagNames,
			TagValues: tagValues,
		})
		if err != nil {
			logger.WithError(err).Errorln(sql)
			errList[pointIndex] = err.Error()
			if !haveError {
				haveError = true
			}
		}
	}
	if haveError {
		return errors.New(strings.Join(errList, ","))
	}
	return nil
}
