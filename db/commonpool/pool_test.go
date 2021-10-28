package commonpool

import (
	"testing"

	"github.com/taosdata/blm3/config"
)

func TestMain(m *testing.M) {
	config.Init()
	m.Run()
}

func BenchmarkGetConnection(b *testing.B) {
	for i := 0; i < b.N; i++ {
		conn, err := GetConnection("root", "taosdata")
		if err != nil {
			b.Error(err)
			return
		}
		conn.Put()
	}
}
