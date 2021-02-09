package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/kataras/golog"
	"github.com/skyaxl/synack-proxy/pkg/proxy"
	"github.com/skyaxl/synack-proxy/pkg/registry"
	"github.com/skyaxl/synack-proxy/pkg/response/factory"
)

func main() {
	portStr := os.Getenv("APP_PORT")
	port, _ := strconv.ParseInt(portStr, 10, 64)
	if port == 0 {
		port = 8080
	}
	p := proxy.NewHandler(&registry.LocalRegistry{}, factory.NewDefault(), http.DefaultClient)
	errs := make(chan error, 2)
	go func() {
		golog.Infof("[Proxy Api] Has started with address localhost:%d\n", port)
		errs <- http.ListenAndServe(fmt.Sprintf(":%d", port), p)
	}()
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	golog.Infof("[Proxy Api] Has stoped. \n%s\n", <-errs)
}
