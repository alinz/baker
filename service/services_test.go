package service_test

import (
	"sync"
	"testing"
	"time"

	"github.com/alinz/baker"
	"github.com/alinz/baker/pkg/endpoint"
	"github.com/alinz/baker/service"
)

type ConfigLoaderFn func(addr endpoint.HTTPAddr) (*baker.Config, error)

var _ service.ConfigLoader = (*ConfigLoaderFn)(nil)

func (cl ConfigLoaderFn) Config(addr endpoint.HTTPAddr) (*baker.Config, error) {
	return cl(addr)
}

type ConsumerFn struct {
	service func(service *baker.Service) error
	close   func(err error)
}

func (c *ConsumerFn) Service(service *baker.Service) error {
	return c.service(service)
}

func (c *ConsumerFn) Close(err error) {
	c.close(err)
}

func TestServices(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	dummyConfigLoader := ConfigLoaderFn(func(addr endpoint.HTTPAddr) (*baker.Config, error) {
		return &baker.Config{
			Domain: "example.com",
			Path:   "/api",
		}, nil
	})

	consumerFn := &ConsumerFn{
		service: func(service *baker.Service) error {
			wg.Done()
			return nil
		},
		close: func(err error) {

		},
	}

	services := service.New(dummyConfigLoader, 1*time.Second)

	go services.Pipe(consumerFn)

	addr := endpoint.NewAddr("0.0.0.0", 8000, false)
	services.Container(&baker.Container{
		ID:       "1",
		Active:   true,
		Addr:     addr,
		PingAddr: endpoint.NewHTTPAddr(addr, "/ping"),
		Err:      nil,
	})

	wg.Wait()
}
