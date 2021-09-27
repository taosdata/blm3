package db

import (
	"fmt"
	"github.com/silenceper/pool"
	"github.com/taosdata/driver-go/v2/af"
	"time"
)

type AFConnectorPool struct {
	host     string
	user     string
	password string
	port     int
	db       string
	pool     pool.Pool
}

func NewAFConnectorPool(host, user, password, db string, port int, maxConnect, maxIdle int, idleTimeout time.Duration) (*AFConnectorPool, error) {
	conn, err := af.Open(host, user, password, "", port)
	if err != nil {
		return nil, err
	}
	if db != "" {
		_, err = conn.Exec(fmt.Sprintf("create database if not exists %s", db))
		if err != nil {
			return nil, err
		}
		err = conn.Close()
		if err != nil {
			return nil, err
		}
	}
	a := &AFConnectorPool{db: db}
	poolConfig := &pool.Config{
		InitialCap:  1,
		MaxCap:      maxConnect,
		MaxIdle:     maxIdle,
		Factory:     a.factory,
		Close:       a.close,
		IdleTimeout: idleTimeout,
	}
	p, err := pool.NewChannelPool(poolConfig)
	if err != nil {
		return nil, err
	}
	a.pool = p
	return a, nil
}

func (a *AFConnectorPool) factory() (interface{}, error) {
	return af.Open(a.host, a.user, a.password, a.db, a.port)
}

func (a *AFConnectorPool) close(v interface{}) error {
	if v != nil {
		return v.(*af.Connector).Close()
	}
	return nil
}

func (a *AFConnectorPool) Get() (*af.Connector, error) {
	v, err := a.pool.Get()
	if err != nil {
		return nil, err
	}
	return v.(*af.Connector), nil
}

func (a *AFConnectorPool) Put(c *af.Connector) error {
	return a.pool.Put(c)
}

func (a *AFConnectorPool) Close(c *af.Connector) error {
	return a.pool.Close(c)
}

func (a *AFConnectorPool) Release() {
	a.pool.Release()
}
