package service_test

import (
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

func TestServices(t *testing.T) {
	dummyConfigLoader := ConfigLoaderFn(func(addr endpoint.HTTPAddr) (*baker.Config, error) {
		return nil, nil
	})

	services := service.New(dummyConfigLoader, 1*time.Second)

	services.Start(nil)
}
