package statsd

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/agent"
	"github.com/influxdata/telegraf/plugins/inputs/statsd"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/sirupsen/logrus"
	dbPackage "github.com/taosdata/blm3/db/advancepool"
	"github.com/taosdata/blm3/log"
	"github.com/taosdata/blm3/plugin"
	"time"
)

var logger = log.GetLogger("statsd")

type Plugin struct {
	conf       Config
	ac         telegraf.Accumulator
	input      *statsd.Statsd
	closeChan  chan struct{}
	metricChan chan telegraf.Metric
	serializer *influx.Serializer
}

func (p *Plugin) Init(r gin.IRouter) error {
	p.conf.setValue()
	if !p.conf.Enable {
		logger.Info("statsd disabled")
		return nil
	}
	p.serializer = influx.NewSerializer()
	p.metricChan = make(chan telegraf.Metric, 2*p.conf.Worker)
	for i := 0; i < p.conf.Worker; i++ {
		go func() {
			for {
				select {
				case metric := <-p.metricChan:
					p.HandleMetrics(metric)
				case <-p.closeChan:
					return
				}
			}
		}()
	}
	p.input = &statsd.Statsd{
		Protocol:               p.conf.Protocol,
		ServiceAddress:         fmt.Sprintf(":%d", p.conf.Port),
		MaxTCPConnections:      p.conf.MaxTCPConnections,
		TCPKeepAlive:           p.conf.TCPKeepAlive,
		AllowedPendingMessages: p.conf.AllowedPendingMessages,
		DeleteCounters:         p.conf.DeleteCounters,
		DeleteGauges:           p.conf.DeleteGauges,
		DeleteSets:             p.conf.DeleteSets,
		DeleteTimings:          p.conf.DeleteTimings,
		Log:                    logger,
	}
	p.ac = agent.NewAccumulator(&MetricMaker{logger: logger}, p.metricChan)
	return nil
}

func (p *Plugin) Start() error {
	if !p.conf.Enable {
		return nil
	}
	err := p.input.Start(p.ac)
	if err != nil {
		return err
	}
	p.closeChan = make(chan struct{})
	ticker := time.NewTicker(p.conf.GatherInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				err := p.input.Gather(p.ac)
				if err != nil {
					logger.WithError(err).Error("gather error")
				}
			case <-p.closeChan:
				p.closeChan = nil
				ticker.Stop()
				ticker = nil
				return
			}
		}

	}()
	return nil
}

func (p *Plugin) Stop() error {
	if !p.conf.Enable {
		return nil
	}
	p.input.Stop()
	close(p.closeChan)
	return nil
}

func (p *Plugin) String() string {
	return "statsd"
}

func (p *Plugin) Version() string {
	return "v1"
}

func (p *Plugin) HandleMetrics(metrics ...telegraf.Metric) {
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

type MetricMaker struct {
	logger logrus.FieldLogger
}

func (m *MetricMaker) LogName() string {
	return "metric"
}

func (m *MetricMaker) MakeMetric(metric telegraf.Metric) telegraf.Metric {
	return metric
}

func (m *MetricMaker) Log() telegraf.Logger {
	return m.logger
}

func init() {
	plugin.Register(&Plugin{})
}
