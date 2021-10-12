package db

import (
	"github.com/huskar-t/blm_demo/config"
	"github.com/silenceper/pool"
	"github.com/taosdata/driver-go/v2/af"
	"github.com/taosdata/driver-go/v2/wrapper"
	"sync"
	"unsafe"
)

type AdvancePool struct {
	user     string
	password string
	pool     pool.Pool
}

func NewAdvancePool(user, password string) (*AdvancePool, error) {
	conn, err := wrapper.TaosConnect("", user, password, "", 0)
	if err != nil {
		return nil, err
	}
	defer wrapper.TaosClose(conn)
	a := &AdvancePool{user: user, password: password}
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

func (a *AdvancePool) factory() (interface{}, error) {
	taos, err := wrapper.TaosConnect("", a.user, a.password, "", 0)
	if err != nil {
		return nil, err
	}
	return af.NewConnector(taos)
}

func (a *AdvancePool) close(v interface{}) error {
	if v != nil {
		return v.(*af.Connector).Close()
	}
	return nil
}

func (a *AdvancePool) Get() (*af.Connector, error) {
	v, err := a.pool.Get()
	if err != nil {
		return nil, err
	}
	return v.(*af.Connector), nil
}

func (a *AdvancePool) Put(c *af.Connector) error {
	return a.pool.Put(c)
}

func (a *AdvancePool) Close(c unsafe.Pointer) error {
	return a.pool.Close(c)
}

func (a *AdvancePool) Release() {
	a.pool.Release()
}

func (a *AdvancePool) verifyPassword(password string) bool {
	return password == a.password
}

var advanceConnectionMap = sync.Map{}

type AdvanceConn struct {
	TaosConnection *af.Connector
	pool           *AdvancePool
}

func (c *AdvanceConn) Put() error {
	return c.pool.Put(c.TaosConnection)
}

func GetAdvanceConnection(user, password string) (*AdvanceConn, error) {
	p, exist := advanceConnectionMap.Load(user)
	if exist {
		connectionPool := p.(*AdvancePool)
		if !connectionPool.verifyPassword(password) {
			newPool, err := NewAdvancePool(user, password)
			if err != nil {
				return nil, err
			}
			connectionPool.Release()
			advanceConnectionMap.Store(user, newPool)
			c, err := newPool.Get()
			if err != nil {
				return nil, err
			}
			return &AdvanceConn{
				TaosConnection: c,
				pool:           newPool,
			}, nil
		} else {
			c, err := connectionPool.Get()
			if err != nil {
				return nil, err
			}
			return &AdvanceConn{
				TaosConnection: c,
				pool:           connectionPool,
			}, nil
		}
	} else {
		newPool, err := NewAdvancePool(user, password)
		if err != nil {
			return nil, err
		}
		connectionMap.Store(user, newPool)
		c, err := newPool.Get()
		if err != nil {
			return nil, err
		}
		return &AdvanceConn{
			TaosConnection: c,
			pool:           newPool,
		}, nil
	}
}
