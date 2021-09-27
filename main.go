package main

import (
	"fmt"
	"github.com/huskar-t/blm_demo/config"
	"github.com/huskar-t/blm_demo/plugin"
	_ "github.com/huskar-t/blm_demo/plugin/opentsdb"
	"github.com/huskar-t/blm_demo/rest"
	"github.com/taosdata/go-utils/web"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	router := web.CreateRouter(config.Conf.Debug, &config.Conf.Cors, false)
	r := rest.Restful{}
	_ = r.Init(router)
	plugin.RegisterGenerateAuth(router)
	plugin.Init(router)
	plugin.Start()
	config.Clear()
	server := &http.Server{
		Addr:              ":" + strconv.Itoa(config.Conf.Port),
		Handler:           router,
		ReadHeaderTimeout: 20 * time.Second,
		ReadTimeout:       200 * time.Second,
		WriteTimeout:      30 * time.Second,
	}
	fmt.Println("server on :", config.Conf.Port)
	var g errgroup.Group
	g.Go(server.ListenAndServe)
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-quit
	fmt.Println("stop server start")
	plugin.Stop()
	fmt.Println("stop server finished")
}
