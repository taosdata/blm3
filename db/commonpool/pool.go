package commonpool

import (
	"container/list"
	"sync"
	"unsafe"

	"github.com/taosdata/blm3/config"
	"github.com/taosdata/blm3/connpool"
)

var connectionMap = sync.Map{}

type Conn struct {
	e              *list.Element
	TaosConnection unsafe.Pointer
	pool           *connpool.Pool
}

func (c *Conn) Put() error {
	return c.pool.Put(c.e)
}

func GetConnection(user, password string) (*Conn, error) {
	p, exist := connectionMap.Load(user)
	if exist {
		connectionPool := p.(*connpool.Pool)
		if !connectionPool.VerifyPassword(password) {
			newPool, err := connpool.NewConnPool(config.Conf.Pool.MaxConnect, config.Conf.Pool.MaxIdle, user, password)
			if err != nil {
				return nil, err
			}
			connectionPool.Release()
			connectionMap.Store(user, newPool)
			c, err := newPool.Get()
			if err != nil {
				return nil, err
			}
			return &Conn{
				e:              c,
				TaosConnection: c.Value.(unsafe.Pointer),
				pool:           newPool,
			}, nil
		} else {
			c, err := connectionPool.Get()
			if err != nil {
				return nil, err
			}
			return &Conn{
				e:              c,
				TaosConnection: c.Value.(unsafe.Pointer),
				pool:           connectionPool,
			}, nil
		}
	} else {
		newPool, err := connpool.NewConnPool(config.Conf.Pool.MaxConnect, config.Conf.Pool.MaxIdle, user, password)
		if err != nil {
			return nil, err
		}
		connectionMap.Store(user, newPool)
		c, err := newPool.Get()
		if err != nil {
			return nil, err
		}
		return &Conn{
			e:              c,
			TaosConnection: c.Value.(unsafe.Pointer),
			pool:           newPool,
		}, nil
	}
}
