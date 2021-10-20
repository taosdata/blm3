package commonpool

import (
	"database/sql/driver"
	"sync"
	"unsafe"

	"github.com/silenceper/pool"
	"github.com/taosdata/blm3/config"
	"github.com/taosdata/driver-go/v2/af"
	"github.com/taosdata/driver-go/v2/wrapper"
)

type ConnectorPool struct {
	user     string
	password string
	pool     pool.Pool
}

func NewConnectorPool(user, password string) (*ConnectorPool, error) {
	conn, err := wrapper.TaosConnect("", user, password, "", 0)
	if err != nil {
		return nil, err
	}
	defer wrapper.TaosClose(conn)
	a := &ConnectorPool{user: user, password: password}
	poolConfig := &pool.Config{
		InitialCap:  1,
		MaxCap:      config.Conf.Pool.MaxConnect,
		MaxIdle:     config.Conf.Pool.MaxIdle,
		Factory:     a.factory,
		Close:       a.close,
		IdleTimeout: config.Conf.Pool.IdleTimeout,
	}
	p, err := pool.NewChannelPool(poolConfig)
	if err != nil {
		return nil, err
	}
	a.pool = p
	return a, nil
}

func (a *ConnectorPool) factory() (interface{}, error) {
	return wrapper.TaosConnect("", a.user, a.password, "", 0)
}

func (a *ConnectorPool) close(v interface{}) error {
	if v != nil {
		wrapper.TaosClose(v.(unsafe.Pointer))
	}
	return nil
}

func (a *ConnectorPool) Get() (unsafe.Pointer, error) {
	v, err := a.pool.Get()
	if err != nil {
		return nil, err
	}
	return v.(unsafe.Pointer), nil
}

func (a *ConnectorPool) Put(c unsafe.Pointer) error {
	conn, _ := af.NewConnector(c)
	rows, err := conn.Query("select database()")
	if err != nil {
		a.pool.Close(c)
		return err
	}
	defer rows.Close()
	values := make([]driver.Value, 1)
	err = rows.Next(values)
	if err != nil {
		a.pool.Close(c)
		return err
	}
	if values[0] != nil {
		return a.pool.Close(c)
	}
	return a.pool.Put(c)
}

func (a *ConnectorPool) Close(c unsafe.Pointer) error {
	return a.pool.Close(c)
}

func (a *ConnectorPool) Release() {
	a.pool.Release()
}

func (a *ConnectorPool) verifyPassword(password string) bool {
	return password == a.password
}

var connectionMap = sync.Map{}

type Conn struct {
	TaosConnection unsafe.Pointer
	pool           *ConnectorPool
}

func (c *Conn) Put() error {
	return c.pool.Put(c.TaosConnection)
}

func GetConnection(user, password string) (*Conn, error) {
	p, exist := connectionMap.Load(user)
	if exist {
		connectionPool := p.(*ConnectorPool)
		if !connectionPool.verifyPassword(password) {
			newPool, err := NewConnectorPool(user, password)
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
				TaosConnection: c,
				pool:           newPool,
			}, nil
		} else {
			c, err := connectionPool.Get()
			if err != nil {
				return nil, err
			}
			return &Conn{
				TaosConnection: c,
				pool:           connectionPool,
			}, nil
		}
	} else {
		newPool, err := NewConnectorPool(user, password)
		if err != nil {
			return nil, err
		}
		connectionMap.Store(user, newPool)
		c, err := newPool.Get()
		if err != nil {
			return nil, err
		}
		return &Conn{
			TaosConnection: c,
			pool:           newPool,
		}, nil
	}
}
