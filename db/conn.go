package db

import (
	"github.com/taosdata/driver-go/v2/af"
	"github.com/taosdata/driver-go/v2/wrapper"
	"sync/atomic"
	"unsafe"
)

var count = int32(0)

func CloseConn(taos unsafe.Pointer) bool {
	if !atomic.CompareAndSwapInt32(&count, 0, 1) {
		wrapper.TaosClose(taos)
		return true
	}
	return false
}

func CloseAfConn(conn *af.Connector) bool {
	if !atomic.CompareAndSwapInt32(&count, 0, 1) {
		conn.Close()
		return true
	}
	return false
}
