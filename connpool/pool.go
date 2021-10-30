package connpool

import (
	"container/list"
	"errors"
	"sync"
	"unsafe"

	"github.com/taosdata/driver-go/v2/wrapper"
)

var TooManyConnectError = errors.New("too many connect")

type Pool struct {
	usingLock  sync.RWMutex
	idleLock   sync.RWMutex
	usingList  *list.List
	idleList   *list.List
	maxConnect int
	maxIdle    int
	user       string
	password   string
}

func NewConnPool(maxConnect, maxIdle int, user, password string) (*Pool, error) {
	p := &Pool{
		usingList:  list.New(),
		idleList:   list.New(),
		maxConnect: maxConnect,
		maxIdle:    maxIdle,
		user:       user,
		password:   password,
	}
	l, err := p.Get()
	if err != nil {
		return nil, err
	}
	err = p.Put(l)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Pool) Get() (*list.Element, error) {
	p.idleLock.Lock()
	front := p.idleList.Front()
	if front != nil {
		conn := p.idleList.Remove(front).(unsafe.Pointer)
		p.idleLock.Unlock()
		p.usingLock.Lock()
		element := p.usingList.PushBack(conn)
		p.usingLock.Unlock()
		return element, nil
	}
	p.usingLock.Lock()
	defer func() {
		p.usingLock.Unlock()
		p.idleLock.Unlock()
	}()
	if p.usingList.Len()+p.idleList.Len() < p.maxConnect {
		var conn unsafe.Pointer
		var err error
		//thread.Lock()
		conn, err = wrapper.TaosConnect("", p.user, p.password, "", 0)
		//thread.Unlock()
		if err != nil {
			return nil, err
		}
		return p.usingList.PushBack(conn), nil
	} else {
		return nil, TooManyConnectError
	}
}

func (p *Pool) Put(e *list.Element) error {
	p.usingLock.Lock()
	taosConnect := p.usingList.Remove(e).(unsafe.Pointer)
	p.usingLock.Unlock()
	p.idleLock.Lock()
	if p.maxIdle > 0 {
		if p.idleList.Len() >= p.maxIdle {
			p.idleLock.Unlock()
			//thread.Lock()
			wrapper.TaosClose(taosConnect)
			//thread.Unlock()
			return nil
		}
		p.usingLock.Lock()
		if p.idleList.Len()+p.usingList.Len() >= p.maxConnect {
			p.idleLock.Unlock()
			p.usingLock.Unlock()
			//thread.Lock()
			wrapper.TaosClose(taosConnect)
			//thread.Unlock()
			return nil
		}
		p.usingLock.Unlock()
	}
	//conn, err := af.NewConnector(taosConnect)
	//if err != nil {
	//	p.idleLock.Unlock()
	//	return err
	//}
	//rows, err := conn.Query("select database()")
	//if err != nil {
	//	p.idleLock.Unlock()
	//	_ = conn.Close()
	//	return err
	//}
	//defer rows.Close()
	//values := make([]driver.Value, 1)
	//err = rows.Next(values)
	//if err != nil {
	//	p.idleLock.Unlock()
	//	_ = conn.Close()
	//	return err
	//}
	//if values[0] != nil {
	//	p.idleLock.Unlock()
	//	_ = conn.Close()
	//	return nil
	//}
	p.idleList.PushBack(taosConnect)
	p.idleLock.Unlock()
	return nil
}

func (p *Pool) Release() {
	p.idleLock.Lock()
	p.usingLock.Lock()
	defer func() {
		p.idleLock.Unlock()
		p.usingLock.Unlock()
	}()
	f := p.idleList.Front()

	for {
		if f != nil {
			v := f.Value.(unsafe.Pointer)
			//thread.Lock()
			wrapper.TaosClose(v)
			//thread.Unlock()
		} else {
			break
		}
		f = f.Next()
	}

	f = p.usingList.Front()
	for {
		if f != nil {
			v := f.Value.(unsafe.Pointer)
			//thread.Lock()
			wrapper.TaosClose(v)
			//thread.Unlock()
		} else {
			break
		}
		f = f.Next()
	}
}

func (p *Pool) VerifyPassword(password string) bool {
	return password == p.password
}
