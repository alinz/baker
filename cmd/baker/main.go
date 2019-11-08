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
	acmeEnable := os.Getenv("ACME") == "true"
	acmePath := os.Getenv("ACME_PATH")

	if acmePath == "" {
		acmePath = "."
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
