package gateway_test

import (
	"fmt"
	"testing"

	"github.com/alinz/baker"
	"github.com/alinz/baker/gateway"
	"github.com/alinz/baker/pkg/endpoint"
)

func dummyService(id string) *baker.Service {
	addr := endpoint.NewAddr("0.0.0.0", 80, false)

	return &baker.Service{
		Container: &baker.Container{
			ID:       id,
			Active:   true,
			Addr:     addr,
			PingAddr: endpoint.NewHTTPAddr(addr, "/test1"),
		},
		Config: &baker.Config{
			Domain: "example.com",
			Path:   "/test",
		},
	}
}

func TestServices(t *testing.T) {
	t.Skip()

	services := gateway.NewServices()

	service := dummyService("1")
	services.Add(service)

	service = dummyService("2")
	services.Add(service)

	service1 := services.Get()
	if service1 == nil {
		t.Fatal("service should be presented")
	}

	service2 := services.Get()
	if service1.Container.ID == service2.Container.ID {
		t.Fatal("services should be different")
	}

	services.Remove(service)
	service = services.Get()
	if service == nil {
		t.Fatal("service should not be presented")
	}

	service = services.Get()
	if service == nil {
		t.Fatal("service should not be presented")
	}

	services.Remove(service)
	service = services.Get()
	if service != nil {
		t.Fatal("service should not be presented")
	}
}

func TestPaths(t *testing.T) {
	paths := gateway.NewPaths()

	service := dummyService("1")

	paths.Add(service)

	services := paths.Services("/test")
	if services == nil {
		t.Fatal("services should not be nil")
	}

	paths.Remove(service)

	services = paths.Services("/test")
	if services != nil {
		t.Fatal("services should be nil")
	}
}

func TestDomains(t *testing.T) {
	t.Skip()

	domains := gateway.NewDomains()

	service := dummyService("1")

	domains.Add(service)
	domains.Add(service)

	fmt.Printf("%v", domains)
}
