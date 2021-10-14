package parse

import (
	"github.com/taosdata/blm3/common"
	"github.com/taosdata/blm3/tools/pool"
	"strings"
	"time"
)

func Repair(buf []byte, precision string) ([]string, int, error) {
	var result []string
	points, err := ParsePointsWithPrecision(buf, time.Now().UTC(), precision)
	if err != nil {
		return nil, 0, err
	}
	for i, p := range points {
		name := p.Name()
		n := repairName(name)
		p.SetName(n)
		tags := p.Tags()
		newTagMap := make(map[string]string)
		for _, tag := range tags {
			newK := repairName(tag.Key)
			newV := repairValue(tag.Value)
			newTagMap[newK] = newV
		}
		newTags := NewTags(newTagMap)
		p.SetTags(newTags)
		columns, err := p.Fields()
		if err != nil {
			return nil, i, err
		}
		for k, v := range columns {
			newK := repairName([]byte(k))
			if newK != k {
				delete(columns, k)
			}
			newV := v
			vv, ok := v.(string)
			if ok {
				newV = repairValue([]byte(vv))
			}
			columns[newK] = newV
		}
		p.(*point).fields = columns.MarshalBinary()
		result = append(result, p.String())
	}
	return result, 0, nil
}

func repairName(s []byte) string {
	result := pool.BytesPoolGet()
	defer pool.BytesPoolPut(result)
	if (s[0] <= 'z' && s[0] >= 'a') || s[0] == '_' {

	} else {
		result.WriteByte('_')
	}
	for i := 0; i < len(s); i++ {
		b := s[i]
		if !checkByte(b) {
			result.WriteByte('_')
		} else {
			result.WriteByte(b)
		}
	}
	r := result.String()
	if isReservedWords(r) {
		result.Reset()
		result.WriteByte('_')
		result.WriteString(r)
		return result.String()
	} else {
		return r
	}
}

func checkByte(b byte) bool {
	if (b <= 'z' && b >= 'a') || b == '_' || (b <= '9' && b >= '0') {
		return true
	}
	return false
}

func isReservedWords(s string) bool {
	if _, ok := common.ReservedWords[strings.ToUpper(s)]; ok {
		return true
	}
	return false
}

func repairValue(b []byte) string {
	result := pool.BytesPoolGet()
	defer pool.BytesPoolPut(result)
	for i := 0; i < len(b); i++ {
		if b[i] == ' ' {
			result.WriteByte('_')
		} else {
			result.WriteByte(b[i])
		}
	}
	return result.String()
}
