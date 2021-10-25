package icinga2

import (
	"crypto/md5"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/taosdata/blm3/db/commonpool"
	"github.com/taosdata/blm3/log"
	"github.com/taosdata/blm3/plugin"
	"github.com/taosdata/blm3/schemaless"
	"github.com/taosdata/blm3/tools/pool"
)

var logger = log.GetLogger("icinga2")

type Icinga2 struct {
	conf     Config
	client   *http.Client
	request  map[string]*http.Request
	baseURL  *url.URL
	exitChan chan struct{}
}

func (p *Icinga2) Init(_ gin.IRouter) error {
	p.conf.setValue()
	if !p.conf.Enable {
		logger.Info("icinga2 disabled")
		return nil
	}
	if p.conf.ResponseTimeout < time.Second {
		p.conf.ResponseTimeout = time.Second * 5
	}
	certPool := x509.NewCertPool()
	if len(p.conf.CaCertFile) != 0 {
		caCert, err := ioutil.ReadFile(p.conf.CaCertFile)
		if err != nil {
			return err
		}
		certPool.AppendCertsFromPEM(caCert)
	}
	var certificates []tls.Certificate
	if len(p.conf.CertFile) != 0 {
		cert, err := tls.LoadX509KeyPair(p.conf.CertFile, p.conf.KeyFile)
		if err != nil {
			return err
		}
		certificates = append(certificates, cert)
	}

	tlsCfg := &tls.Config{
		RootCAs:            certPool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: p.conf.InsecureSkipVerify,
		Certificates:       certificates,
	}
	p.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: p.conf.ResponseTimeout,
	}
	//hosts
	hostRequestURL := fmt.Sprintf("%s/v1/objects/hosts?attrs=name&attrs=display_name&attrs=state&attrs=check_command", p.conf.Host)
	hostReq, err := p.prepareRequest(hostRequestURL)
	if err != nil {
		return err
	}
	p.request = make(map[string]*http.Request, 2)
	p.request["hosts"] = hostReq
	//services
	serviceRequestURL := fmt.Sprintf("%s/v1/objects/services?attrs=name&attrs=display_name&attrs=state&attrs=check_command&attrs=host_name", p.conf.Host)
	serviceReq, err := p.prepareRequest(serviceRequestURL)
	if err != nil {
		return err
	}
	p.request["services"] = serviceReq
	p.baseURL, err = url.Parse(p.conf.Host)
	if err != nil {
		return err
	}
	return nil
}

func (p *Icinga2) prepareRequest(u string) (*http.Request, error) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	if p.conf.HttpUsername != "" {
		req.SetBasicAuth(p.conf.HttpUsername, p.conf.HttpPassword)
	}
	return req, nil
}

func (p *Icinga2) Start() error {
	p.exitChan = make(chan struct{})
	ticker := time.NewTicker(p.conf.GatherDuration)
	go func() {
		for true {
			select {
			case <-ticker.C:
				err := p.Gather()
				if err != nil {
					logger.WithError(err).Warnln("gather error")
				}
			case <-p.exitChan:
				ticker.Stop()
				ticker = nil
				return
			}
		}
	}()
	return nil
}

func (p *Icinga2) Stop() error {
	if p.exitChan != nil {
		close(p.exitChan)
	}
	return nil
}

func (p *Icinga2) String() string {
	return "icinga2"
}

func (p *Icinga2) Version() string {
	return "v1"
}

type Result struct {
	Results []Object `json:"results"`
}

type Object struct {
	Attrs Attribute `json:"attrs"`
	Name  string    `json:"name"`
	Joins struct{}  `json:"joins"`
	Meta  struct{}  `json:"meta"`
	Type  string    `json:"type"`
}

type Attribute struct {
	CheckCommand string  `json:"check_command"`
	DisplayName  string  `json:"display_name"`
	Name         string  `json:"name"`
	State        float64 `json:"state"`
	HostName     string  `json:"host_name"`
}

func (p *Icinga2) Gather() error {
	var errorList []string
	for objectType, req := range p.request {
		err := p.GatherStatus(objectType, req)
		if err != nil {
			errorList = append(errorList, err.Error())
		}
	}
	if len(errorList) != 0 {
		return errors.New(strings.Join(errorList, ","))
	}
	return nil
}

var levels = []string{"ok", "warning", "critical", "unknown"}
var tagKeys = []string{"check_command", "display_name", "port", "scheme", "server", "source", "state"}

func (p *Icinga2) GatherStatus(objectType string, req *http.Request) error {
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		d, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("%s got status code %d,body:%s\n", req.URL.RequestURI(), resp.StatusCode, string(d))
	}
	result := Result{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}

	conn, err := commonpool.GetConnection(p.conf.User, p.conf.Password)
	if err != nil {
		logger.WithError(err).Errorln("commonpool.GetConnection error")
		return err
	}
	defer conn.Put()
	e, err := schemaless.NewExecutor(conn.TaosConnection)
	if err != nil {
		logger.WithError(err).Errorln("schemalee.NewExecutor error")
		return err
	}
	now := time.Now()
	b := pool.BytesPoolGet()
	defer pool.BytesPoolPut(b)
	for _, check := range result.Results {
		state := int64(check.Attrs.State)
		fields := map[string]interface{}{
			"name":       check.Attrs.Name,
			"state_code": state,
		}
		// source is dependent on 'services' or 'hosts' check
		source := check.Attrs.Name
		if objectType == "services" {
			source = check.Attrs.HostName
		}
		tagValues := []string{
			check.Attrs.CheckCommand,
			check.Attrs.DisplayName,
			p.baseURL.Port(),
			p.baseURL.Scheme,
			p.baseURL.Hostname(),
			source,
			levels[state],
		}
		b.Reset()
		b.WriteString(objectType)
		b.WriteByte(' ')
		for i := 0; i < 7; i++ {
			b.WriteString(tagKeys[i])
			b.WriteByte('=')
			b.WriteString(tagValues[i])
			if i != 6 {
				b.WriteByte(' ')
			}
		}
		tableName := fmt.Sprintf("_%x", md5.Sum(b.Bytes()))
		b.Reset()
		b.WriteByte('\'')
		b.WriteString(objectType)
		b.WriteByte('\'')
		sql, err := e.InsertTDengine(&schemaless.InsertLine{
			DB:         p.conf.DB,
			Ts:         now,
			TableName:  tableName,
			STableName: b.String(),
			Fields:     fields,
			TagNames:   tagKeys,
			TagValues:  tagValues,
		})
		if err != nil {
			logger.WithError(err).Errorln(sql)
		}
	}
	return nil
}

func init() {
	plugin.Register(&Icinga2{})
}
