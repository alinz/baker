package main

import (
	"net/http"
	"os"
	"time"

	"github.com/alinz/baker/container"
	"github.com/alinz/baker/gateway"
	"github.com/alinz/baker/pkg/acme"
	"github.com/alinz/baker/pkg/logger"
	"github.com/alinz/baker/service"
)

func main() {
	acmeEnable := os.Getenv("BAKER_ACME") == "true"
	acmePath := os.Getenv("BAKER_ACME_PATH")
	debugLevel := os.Getenv("BAKER_DEBUG_LEVEL") == "true"

	if acmePath == "" {
		acmePath = "."
	}

	if debugLevel {
		logger.Level = logger.DEBUG_LEVEL
	}

	proxy := gateway.NewHandler()

	containerProducer := container.NewDocker(container.DefaultClient, container.DefaultAddr)
	serviceProducer := service.New(service.NewConfigLoader(nil), 10*time.Second)

	// container -> service producer -> service
	go containerProducer.Pipe(serviceProducer)
	// service -> proxy -> create a map
	go serviceProducer.Pipe(proxy)

	if !acmeEnable {
		if err := http.ListenAndServe(":80", proxy); err != nil {
			logger.Error(err.Error())
		}
		return
	}

	server := acme.NewServer(proxy, acmePath)
	if err := server.Start(proxy); err != nil {
		logger.Error(err.Error())
	}
}
