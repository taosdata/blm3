package collectd

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers/collectd"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/spf13/viper"
	"github.com/taosdata/blm3/db/advancedpool"
	"github.com/taosdata/blm3/log"
	"github.com/taosdata/blm3/plugin"
	"github.com/taosdata/blm3/tools/influxdb/parse"
	"net"
	"strings"
	"time"
)

var logger = log.GetLogger("collectd")

type Plugin struct {
	conf       Config
	conn       net.PacketConn
	serializer *influx.Serializer
	parser     *collectd.CollectdParser
	metricChan chan []telegraf.Metric
	closeChan  chan struct{}
}

func (p *Plugin) Init(_ gin.IRouter) error {
	p.conf.setValue()
	if !p.conf.Enable {
		logger.Info("collectd disabled")
		return nil
	}
	p.conf.Port = viper.GetInt("collectd.port")
	p.conf.DB = viper.GetString("collectd.db")
	p.conf.User = viper.GetString("collectd.user")
	p.conf.Password = viper.GetString("collectd.password")
	p.parser = &collectd.CollectdParser{
		ParseMultiValue: "split",
	}
	return nil
}

func (p *Plugin) Start() error {
	if p.conn != nil {
		err := p.conn.Close()
		if err != nil {
			return err
		}
	}
	conn, err := udpListen("udp", fmt.Sprintf(":%d", p.conf.Port))
	if err != nil {
		return err
	}
	p.closeChan = make(chan struct{})
	p.metricChan = make(chan []telegraf.Metric, 2*p.conf.Worker)
	for i := 0; i < p.conf.Worker; i++ {
		go func() {
			serializer := influx.NewSerializer()
			for {
				select {
				case metric := <-p.metricChan:
					p.HandleMetrics(serializer, metric)
				case <-p.closeChan:
					return
				}
			}
		}()
	}
	p.conn = conn
	go p.listen()
	return nil
}

func (p *Plugin) Stop() error {
	if !p.conf.Enable {
		return nil
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	close(p.closeChan)
	return nil
}

func (p *Plugin) String() string {
	return "collectd"
}

func (p *Plugin) Version() string {
	return "v1"
}

func (p *Plugin) HandleMetrics(serializer *influx.Serializer, metrics []telegraf.Metric) {
	if len(metrics) == 0 {
		return
	}
	lines := make([]string, 0, len(metrics))
	for _, metric := range metrics {
		data, err := serializer.Serialize(metric)
		if err != nil {
			logger.WithError(err).Error("serialize collectd error")
			continue
		}
		l, wrongIndex, err := parse.Repair(data, "ns")
		if err != nil {
			logger.WithError(err).Error("serialize collectd error", l, wrongIndex)
			continue
		}
		lines = append(lines, l...)
	}
	taosConn, err := advancedpool.GetAdvanceConnection(p.conf.User, p.conf.Password)
	if err != nil {
		logger.WithError(err).Errorln("connect taosd error")
		return
	}
	defer func() {
		putErr := taosConn.Put()
		if putErr != nil {
			logger.WithError(putErr).Errorln("taos connect pool put error")
		}
	}()
	conn := taosConn.TaosConnection
	_, err = conn.Exec(fmt.Sprintf("create database if not exists %s precision 'ns' update 2", p.conf.DB))
	if err != nil {
		logger.WithError(err).Errorln("create database error", p.conf.DB)
		return
	}
	_, err = conn.Exec(fmt.Sprintf("use %s", p.conf.DB))
	if err != nil {
		logger.WithError(err).Error("change to database error", p.conf.DB)
		return
	}
	start := time.Now()
	logger.Debugln(start, "insert lines", lines)
	err = conn.InfluxDBInsertLines(lines, "ns")
	logger.Debugln("insert lines finish cost:", time.Now().Sub(start), lines)
	if err != nil {
		logger.WithError(err).Errorln("insert lines error", lines)
		return
	}
}

func (p *Plugin) listen() {
	buf := make([]byte, 64*1024) // 64kb - maximum size of IP packet
	for {
		n, _, err := p.conn.ReadFrom(buf)
		if err != nil {
			if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
				logger.Error(err.Error())
			}
			break
		}

		metrics, err := p.parser.Parse(buf[:n])
		if err != nil {
			logger.Errorf("Unable to parse incoming packet: %s", err.Error())
			continue
		}
		p.metricChan <- metrics

	}
}

func udpListen(network string, address string) (net.PacketConn, error) {
	switch network {
	case "udp", "udp4", "udp6":
		var addr *net.UDPAddr
		var err error
		var ifi *net.Interface
		if spl := strings.SplitN(address, "%", 2); len(spl) == 2 {
			address = spl[0]
			ifi, err = net.InterfaceByName(spl[1])
			if err != nil {
				return nil, err
			}
		}
		addr, err = net.ResolveUDPAddr(network, address)
		if err != nil {
			return nil, err
		}
		if addr.IP.IsMulticast() {
			return net.ListenMulticastUDP(network, ifi, addr)
		}
		return net.ListenUDP(network, addr)
	}
	return net.ListenPacket(network, address)
}

func init() {
	plugin.Register(&Plugin{})
}
