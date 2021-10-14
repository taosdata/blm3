package collectd

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers/collectd"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/spf13/viper"
	dbPackage "github.com/taosdata/blm3/db/advancepool"
	"github.com/taosdata/blm3/log"
	"github.com/taosdata/blm3/plugin"
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
}

func (p *Plugin) Init(r gin.IRouter) error {
	p.conf.setValue()
	if !p.conf.Enable {
		logger.Info("collectd disabled")
		return nil
	}
	p.conf.Port = viper.GetInt("collectd.port")
	p.conf.DB = viper.GetString("collectd.db")
	p.conf.User = viper.GetString("collectd.user")
	p.conf.Password = viper.GetString("collectd.password")
	p.serializer = influx.NewSerializer()
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
	p.conn = conn
	go p.listen()
	return nil
}

func (p *Plugin) Stop() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

func (p *Plugin) String() string {
	return "collectd"
}

func (p *Plugin) Version() string {
	return "v1"
}

func (p *Plugin) writeMetric(metrics []telegraf.Metric) {
	lines := make([]string, 0, len(metrics))
	for _, metric := range metrics {
		data, err := p.serializer.Serialize(metric)
		if err != nil {
			logger.WithError(err).Error("serialize collectd error")
			continue
		}
		lines = append(lines, string(data[:len(data)-1]))
		taosConn, err := dbPackage.GetAdvanceConnection(p.conf.User, p.conf.Password)
		if err != nil {
			logger.WithError(err).Errorln("connect taosd error")
			return
		}
		defer taosConn.Put()
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
		logger.Debugln("insert lines finish cast:", time.Now().Sub(start))
		if err != nil {
			logger.WithError(err).Errorln("insert lines error", lines)
			return
		}
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
		p.writeMetric(metrics)
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
