//
// Copyright (c) 2017
// Cavium
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package main

import (
	"flag"
	"fmt"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/export/distro"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
)

func main() {
	var useConsul bool
	var useProfile string
	start := time.Now()
	fmt.Println("main fn called")
	fmt.Println("UseConsul is %X", useConsul)
	flag.BoolVar(&useConsul, "consul", false, "Indicates the service should use consul.")
	flag.BoolVar(&useConsul, "c", false, "Indicates the service should use consul.")
	flag.StringVar(&useProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "p", "", "Specify a profile other than default.")
	flag.Usage = usage.HelpCallback
	flag.Parse()
	fmt.Println("UseConsul after flag parse ")

	params := startup.BootParams{UseConsul: useConsul, UseProfile: useProfile, BootTimeout: internal.BootTimeoutDefault}
	fmt.Println("UseConsul is %X", useConsul)
	fmt.Println("bootstrap fn called")
	startup.Bootstrap(params, distro.Retry, logBeforeInit)

	if ok := distro.Init(useConsul); !ok {
		logBeforeInit(fmt.Errorf("%s: Service bootstrap failed", internal.ExportDistroServiceKey))
		return
	}

	fmt.Println("Service dependencies resolved...")
	distro.LoggingClient.Info("Service dependencies resolved...")
	distro.LoggingClient.Info(fmt.Sprintf("Starting %s %s ", internal.ExportDistroServiceKey, edgex.Version))

	http.TimeoutHandler(nil, time.Millisecond*time.Duration(distro.Configuration.Service.Timeout), "Request timed out")
	distro.LoggingClient.Info(distro.Configuration.Service.StartupMsg, "")

	errs := make(chan error, 2)
	eventCh := make(chan *models.Event, 10)

	listenForInterrupt(errs)

	// There can be another receivers that can be initialized here
	fmt.Println("calling Zero MQ Receiver")
	distro.ZeroMQReceiver(eventCh)
	fmt.Println("calling loop")
	distro.Loop(errs, eventCh)

	// Time it took to start service
	distro.LoggingClient.Info("Service started in: "+time.Since(start).String(), "")
	distro.LoggingClient.Info("Listening on port: " + strconv.Itoa(distro.Configuration.Service.Port))
	c := <-errs
	distro.Destruct()
	distro.LoggingClient.Warn(fmt.Sprintf("terminating: %v", c))
}

func logBeforeInit(err error) {
	l := logger.NewClient(internal.ExportDistroServiceKey, false, "", logger.InfoLog)
	l.Error(err.Error())
}

func listenForInterrupt(errChan chan error) {
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}