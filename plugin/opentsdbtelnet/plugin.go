package opentsdbtelnet

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/taosdata/blm3/db/commonpool"
	"github.com/taosdata/blm3/log"
	"github.com/taosdata/blm3/plugin"
	"github.com/taosdata/blm3/schemaless/opentsdb"
)

var logger = log.GetLogger("opentsdb_telnet")

type Plugin struct {
	conf        Config
	done        chan struct{}
	id          uint64
	accept      chan bool
	in          chan []byte
	wg          sync.WaitGroup
	cleanup     sync.Mutex
	TCPListener *net.TCPListener
	connList    map[uint64]*net.TCPConn
}

func (p *Plugin) Init(_ gin.IRouter) error {
	p.conf.setValue()
	if !p.conf.Enable {
		logger.Info("opentsdb_telnet disabled")
		return nil
	}
	p.accept = make(chan bool, p.conf.MaxTCPConnections)
	p.connList = make(map[uint64]*net.TCPConn)
	for i := 0; i < p.conf.MaxTCPConnections; i++ {
		p.accept <- true
	}
	p.in = make(chan []byte, p.conf.Worker*2)
	p.done = make(chan struct{})

	return nil
}

func (p *Plugin) Start() error {
	if !p.conf.Enable {
		return nil
	}
	err := p.tcp(p.conf.Port)
	if err != nil {
		return err
	}
	return nil
}

func (p *Plugin) Stop() error {
	logger.Infof("Stopping the statsd service")
	close(p.done)
	p.TCPListener.Close()
	var tcpConnList []*net.TCPConn
	p.cleanup.Lock()
	for _, conn := range p.connList {
		tcpConnList = append(tcpConnList, conn)
	}
	p.cleanup.Unlock()
	for _, conn := range tcpConnList {
		conn.Close()
	}

	p.wg.Wait()
	close(p.in)
	return nil
}

func (p *Plugin) String() string {
	return "opentsdb_telnet"
}

func (p *Plugin) Version() string {
	return "v1"
}

func (p *Plugin) handler(conn *net.TCPConn, id uint64) {
	defer func() {
		p.wg.Done()
		conn.Close()
		p.accept <- true
		p.forget(id)
	}()
	buffer := make([]byte, 0, 1024)
	d := make([]byte, 1024)
	for {
		select {
		case <-p.done:
			return
		default:
			n, err := conn.Read(d)
			if err != nil {
				fmt.Println(err)
				return
			}
			if n == 0 {
				continue
			}
			buffer = append(buffer, d[:n]...)
			customIndex := 0
			for i := len(buffer) - 1; i > 0; i-- {
				if buffer[i] == '\n' {
					customIndex = i
					break
				}
			}
			if customIndex > 0 {
				data := make([]byte, customIndex)
				copy(data, buffer[:customIndex])
				buffer = buffer[customIndex+1:]
				select {
				case p.in <- data:
				default:
					logger.Errorln("can not handle more message so far. increase opentsdb_telnet.worker")
				}
			}
		}
	}
}

func (p *Plugin) tcp(port int) error {
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		return err
	}

	logger.Infof("TCP listening on %q", listener.Addr().String())
	p.TCPListener = listener
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		if err := p.tcpListen(listener); err != nil {
			logger.WithError(err).Panic()
		}
	}()

	for i := 1; i <= p.conf.Worker; i++ {
		// Start the line parser
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.parser()
		}()
	}
	return nil
}

func (p *Plugin) tcpListen(listener *net.TCPListener) error {
	for {
		select {
		case <-p.done:
			return nil
		default:
			// Accept connection:
			conn, err := listener.AcceptTCP()
			if err != nil {
				return err
			}

			if p.conf.TCPKeepAlive {
				if err = conn.SetKeepAlive(true); err != nil {
					return err
				}
			}

			select {
			case <-p.accept:
				p.wg.Add(1)
				id := atomic.AddUint64(&p.id, 1)
				p.remember(id, conn)
				go p.handler(conn, id)
			default:
				p.refuser(conn)
			}
		}
	}
}

func (p *Plugin) forget(id uint64) {
	p.cleanup.Lock()
	defer p.cleanup.Unlock()
	delete(p.connList, id)
}

func (p *Plugin) remember(id uint64, conn *net.TCPConn) {
	p.cleanup.Lock()
	defer p.cleanup.Unlock()
	p.connList[id] = conn
}

func (p *Plugin) refuser(conn *net.TCPConn) {
	_ = conn.Close()
	logger.Infof("Refused TCP Connection from %s", conn.RemoteAddr())
	logger.Warn("Maximum TCP Connections reached")
}

func (p *Plugin) parser() {
	for {
		select {
		case <-p.done:
			return
		case in := <-p.in:
			p.handleData(in)
		}
	}
}

func (p *Plugin) handleData(data []byte) {
	lines := strings.Split(string(data), "\n")
	taosConn, err := commonpool.GetConnection(p.conf.User, p.conf.Password)
	if err != nil {
		logger.WithError(err).Error("connect taosd error")
		return
	}
	defer func() {
		putErr := taosConn.Put()
		if putErr != nil {
			logger.WithError(putErr).Errorln("taos connect pool put error")
		}
	}()
	var start time.Time
	if logger.Logger.IsLevelEnabled(logrus.DebugLevel) {
		start = time.Now()
	}
	logger.Debug(start, "insert telnet payload", lines)
	var errorList = make([]string, 0, len(lines))
	for _, line := range lines {
		err := opentsdb.InsertTelnet(taosConn.TaosConnection, line, p.conf.DB)
		if err != nil {
			errorList = append(errorList, err.Error())
		}
	}
	logger.Debug("insert telnet payload cost:", time.Now().Sub(start))
	if len(errorList) != 0 {
		logger.WithError(errors.New(strings.Join(errorList, ","))).Error("insert telnet payload error", lines)
		return
	}
}

func init() {
	plugin.Register(&Plugin{})
}
