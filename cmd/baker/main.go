package main

import (
	"net/http"
	"time"

	"github.com/alinz/baker/container"
	"github.com/alinz/baker/gateway"
	"github.com/alinz/baker/service"
)

func main() {
	proxy := gateway.NewHandler()

	containerProducer := container.NewDocker(container.DefaultClient, container.DefaultAddr)
	serviceProducer := service.New(service.NewConfigLoader(nil), 10*time.Second)

	// container -> service producer -> service
	go containerProducer.Start(serviceProducer)
	// service -> proxy -> create a map
	go serviceProducer.Start(proxy)

	http.ListenAndServe(":80", proxy)
}
