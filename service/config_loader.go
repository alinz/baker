package service

import (
	"crypto/tls"
	"encoding/json"
	"net/http"

	"github.com/alinz/bake"
	"github.com/alinz/bake/pkg/endpoint"
)

type ConfigLoader interface {
	Config(addr endpoint.HTTPAddr) (*bake.Config, error)
}

type LoadConfig struct {
	client       *http.Client
	secureClient *http.Client
}

var _ ConfigLoader = (*LoadConfig)(nil)

// Config loads Config object from given address
func (c *LoadConfig) Config(addr endpoint.HTTPAddr) (*bake.Config, error) {
	client := c.client
	if addr.Secure() {
		client = c.secureClient
	}

	// send ping request
	resp, err := client.Get(addr.String())
	if err != nil {
		return nil, err
	}

	// decode ping response
	config := &bake.Config{}
	err = json.NewDecoder(resp.Body).Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func NewConfigLoader(tls *tls.Config) *LoadConfig {
	return &LoadConfig{
		client:       endpoint.NewClient(nil),
		secureClient: endpoint.NewClient(tls),
	}
}
