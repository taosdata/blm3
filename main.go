package main

import (
	"context"
	"fmt"
	"github.com/huskar-t/blm_demo/config"
	"github.com/huskar-t/blm_demo/plugin"
	_ "github.com/huskar-t/blm_demo/plugin/opentsdb"
	"github.com/huskar-t/blm_demo/rest"
	"github.com/taosdata/go-utils/web"
	"log"
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
	go func() {
		// 服务连接
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-quit
	log.Println("Shutdown WebServer ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go func() {
		if err := server.Shutdown(ctx); err != nil {
			log.Println("WebServer Shutdown error:", err)
		}
	}()
	log.Println("Stop Plugins ...")
	ticker := time.NewTicker(time.Second * 5)
	done := make(chan struct{})
	go func() {
		plugin.Stop()
		close(done)
	}()
	select {
	case <-done:
		break
	case <-ticker.C:
		break
	}
	log.Println("Server exiting")
}
